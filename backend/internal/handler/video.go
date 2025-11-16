package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/broker"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

type VideoHandler struct {
	video  interfaces.Video
	action interfaces.Action
	config *config.Config
	s3     interfaces.ObjectStorage
	redis  interfaces.Cache
	kafka  interfaces.MessageBroker
}

func (h *VideoHandler) Upload(c *gin.Context) {
	user, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found in context"})
		return
	}
	u, ok := user.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user type"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	name := c.DefaultQuery("name", strings.Trim(header.Filename, "."+ext))
	allowed := map[string]bool{
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed format video"})
		return
	}
	videoId := uuid.New()
	if err := h.video.Insert(&database.Video{
		VideoId: videoId,
		UserId:  u,
		Name:    name,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tmpInputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input%s", videoId, ext))

	out, err := os.Create(tmpInputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmpInputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out.Close()

	s3Key := fmt.Sprintf("%s/tmp%s", videoId, ext)

	tmpFile, err := os.Open(tmpInputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tmpFile.Close()

	if err := h.s3.PutObject(s3Key, tmpFile); err != nil {
		os.Remove(tmpInputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	os.Remove(tmpInputPath)

	msg := broker.VideoProcessor{
		VideoID: videoId,
		FileExt: ext,
		Done:    false,
	}

	if err := h.kafka.Write(&msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video is uploading",
	})
}

func (h *VideoHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	videoCache := h.redis.Get(id.String())
	if videoCache != "" {
		c.JSON(http.StatusOK, videoCache)
		h.redis.Incr(fmt.Sprintf("views:%s", id))
		return
	}
	video, err := h.video.GetByID(id, h.s3.GetURL)
	if video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	h.redis.Set(id.String(), video, time.Minute*2)
	h.redis.Incr(fmt.Sprintf("views:%s", id))
	userValue, exists := c.Get("user_id")
	if exists {
		user, ok := userValue.(uuid.UUID)
		if ok {
			err := h.action.Insert(user, id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
	c.JSON(http.StatusOK, video)
}

func (h *VideoHandler) GetAll(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	videosCache := h.redis.Get("videos_" + limit + "_" + offset)
	if videosCache != "" {
		c.JSON(http.StatusOK, videosCache)
		return
	}
	videos, err := h.video.GetAll(limit, offset, h.s3.GetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.redis.Set("videos_"+limit+"_"+offset, videos, time.Minute*10)
	c.JSON(http.StatusOK, videos)
}
func (h *VideoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.video.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = h.s3.DeleteObject(id.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
}

func (h *VideoHandler) Upscale(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	real, err := strconv.ParseBool(c.DefaultQuery("real", "false"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	video, err := h.video.GetByID(id, h.s3.GetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if video.Quality >= 1440 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Video too big for upscaling"})
		return
	}
	msg := broker.VideoUpscale{
		VideoID:   id,
		Quality:   video.Quality,
		Realistic: real,
		Done:      false,
	}

	if err := h.kafka.Write(&msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
func (h *VideoHandler) Subtitles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	lang := c.DefaultQuery("lang", "")
	if lang != "" && h.s3.ExitsObjects(fmt.Sprintf("%s/%s_sub.vtt", id, lang)) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s/%s_sub.vtt", h.s3.GetURL(id), lang)})
		return
	}

	msg := broker.VideoSubtitle{
		VideoID:  id,
		Language: lang,
		Done:     false,
	}
	if err := h.kafka.Write(&msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
