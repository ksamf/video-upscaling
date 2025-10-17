package main

import (
	"context"
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
	"github.com/ksamf/video-upscaling/backend/internal/database"
	broker "github.com/ksamf/video-upscaling/backend/internal/kafka"
	"github.com/ksamf/video-upscaling/backend/internal/rest"
)

func (app *application) uploadVideo(c *gin.Context) {
	upscale, _ := strconv.ParseBool(c.DefaultQuery("up", "false"))
	realisticVideo, _ := strconv.ParseBool(c.DefaultQuery("real", "true"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error get file"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	name := c.DefaultQuery("name", strings.Trim(header.Filename, "."+ext))
	allowed := map[string]bool{
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат файла"})
		return
	}

	videoId := uuid.New()
	tmpInputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input%s", videoId, ext))

	out, err := os.Create(tmpInputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tmp file"})
		return
	}

	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmpInputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tmp file"})
		return
	}
	out.Close()

	s3Key := fmt.Sprintf("%s/tmp%s", videoId, ext)

	tmpFile, err := os.Open(tmpInputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open tmp file"})
		return
	}
	defer tmpFile.Close()

	if err := app.s3.PutObject(s3Key, tmpFile); err != nil {
		os.Remove(tmpInputPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to S3"})
		return
	}
	os.Remove(tmpInputPath)

	msg := broker.VideoJob{
		VideoID:        videoId,
		FileName:       name,
		FileExt:        ext,
		Upscale:        upscale,
		RealisticVideo: realisticVideo,
		BaseURL:        app.config.Api.BaseURL,
	}

	if err := broker.Publish(context.Background(), app.kafka.Writer, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish kafka"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Видео успешно загружено",
	})
}

func (app *application) getVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
	}
	videoCache, err := app.redis.Get(c, id.String()).Result()
	if err == nil {
		c.JSON(http.StatusOK, videoCache)
		return
	}
	video, err := app.models.Videos.GetByID(id, app.s3.GetURL)
	if video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video"})
	}
	app.redis.Set(c, id.String(), video, time.Minute*10)
	c.JSON(http.StatusOK, video)
}

func (app *application) getAllVideos(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	videosCache, err := app.redis.Get(c, "videos_"+limit+"_"+offset).Result()
	if err == nil {
		c.JSON(http.StatusOK, videosCache)
		return
	}
	videos, err := app.models.Videos.GetAll(limit, offset, app.s3.GetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get videos: %w", err)})
		return
	}
	app.redis.Set(c, "videos_"+limit+"_"+offset, videos, time.Minute*10)
	c.JSON(http.StatusOK, videos)
}
func (app *application) deleteVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	err = app.models.Videos.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from database"})
		return
	}
	err = app.s3.DeleteObject(id.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video from S3"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
}
func (app *application) updateVideoPartial(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	var updateVideo database.Video
	if err := c.BindJSON(&updateVideo); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updateVideo.VideoId = id
	if updateVideo.LanguageId != 0 {
		err = app.models.Videos.UpdatePartial(id, "language_id", updateVideo.LanguageId)
	}
	if updateVideo.Quality != 0 {
		err = app.models.Videos.UpdatePartial(id, "qualities_id", updateVideo.Quality)
	}
	if updateVideo.Name != "" {
		err = app.models.Videos.UpdatePartial(id, "name", updateVideo.Name)
	}
	if updateVideo.VideoPath != "" {
		err = app.models.Videos.UpdatePartial(id, "video_path", updateVideo.VideoPath)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video updated successfully"})
}

func (app *application) getVideoSubtitles(c *gin.Context) {
	id := c.Param("id")
	lang := c.DefaultQuery("lang", "en")
	subCache, err := app.redis.Get(c, id+"_"+lang+"_sub").Result()
	if err == nil {
		c.JSON(http.StatusOK, subCache)
		return
	}
	if app.s3.ExitsObjects(id + "/" + lang + "_sub.vtt") {
		c.String(http.StatusOK, fmt.Sprintf("https://%s/%s/%s/%s_sub.vtt", app.s3.Endpoint, app.s3.BucketName, id, lang))
	} else {
		err := rest.TranslateSubtitles(uuid.MustParse(id), app.config.Api.BaseURL, lang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate subtitles"})
			return
		}
		subPath := fmt.Sprintf("https://%s/%s/%s/%s_sub.vtt", app.s3.Endpoint, app.s3.BucketName, id, lang)
		app.redis.Set(c, id+"_"+lang+"_sub", subPath, time.Minute*30)
		c.JSON(http.StatusOK, gin.H{"message": subPath})
	}
}

// func (app *application) getVideoDubbing(c *gin.Context) {
// 	id := c.Param("id")
// 	lang := c.DefaultQuery("lang", "en")
// 	if app.s3.ExitsObjects(id + "/" + lang + "_dub.mp3") {
// 		c.String(http.StatusOK, fmt.Sprintf("https://%s/%s/%s/%s_dub.mp3", app.s3.Endpoint, app.s3.BucketName, id, lang))
// 	} else {
// 		if !app.s3.ExitsObjects(id + "/" + lang + "_sub.vtt") {
// 			err := rest.TranslateSubtitles(uuid.MustParse(id), app.config.Api.BaseURL, lang)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate subtitles"})
// 				return
// 			}
// 		}
// 		err := rest.CreateDubbing(uuid.MustParse(id), app.config.Api.BaseURL, lang)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate subtitles"})
// 			return
// 		}
// 		c.String(http.StatusOK, fmt.Sprintf("https://%s/%s/%s/%s_dub.mp3", app.s3.Endpoint, app.s3.BucketName, id, lang))
// 	}
// }
