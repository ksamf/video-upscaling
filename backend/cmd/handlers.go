package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/rest"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
)

func (app *application) uploadVideo(c *gin.Context) {

	upscale := c.DefaultQuery("up", "false")
	realisticVideo := c.DefaultQuery("real", "true")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error get file"})
		return
	}
	name := c.DefaultQuery("name", header.Filename)
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат файла"})
		return
	}

	if err = utils.VideoProcessor(file, name, upscale, realisticVideo, app.config.Api.BaseURL, app.models.Videos, app.s3); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed transcode video"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Видео успешно загружено"})
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
	video, err := app.models.Videos.GetByID(id)
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
	videos, err := app.models.Videos.GetAll(limit, offset)
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
