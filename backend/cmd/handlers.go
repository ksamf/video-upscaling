package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
)

func (app *application) uploadVideo(c *gin.Context) {
	realisticVideo := c.DefaultQuery("realistic", "true")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error get file: %v", err)
		return
	}
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
	}
	if !allowed[ext] {
		c.String(http.StatusBadRequest, "Недопустимый формат файла")
		return
	}

	if err = utils.VideoProcessor(file, header.Filename, realisticVideo, app.models.Videos, app.s3); err != nil {
		c.String(http.StatusInternalServerError, "Failed transcode video: %v", err)
	}

	c.String(http.StatusOK, fmt.Sprint("Видео успешно загружено"))
}

func (app *application) getVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
	}
	video, err := app.models.Videos.GetByID(id)
	if video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video"})
	}
	c.JSON(http.StatusOK, video)
}

func (app *application) getAllVideos(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")
	videos, err := app.models.Videos.GetAll(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}
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
	// if updateVideo.LanguageId != 0 {
	// 	err = app.models.Videos.UpdatePartial(id, "language", updateVideo.LanguageId)
	// }
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

// func (app *application) getVideoSubtitles(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
// 		return
// 	}
// 	subtitles, err := app.models.Videos.GetSubtitles(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video subtitles"})
// 		return
// 	}
// 	if subtitles == nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Subtitles not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, subtitles)
// }

// func (app *application) getVideoDubbing(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
// 		return
// 	}
// 	dubbing, err := app.models.Videos.GetDubbing(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video dubbing"})
// 		return
// 	}
// 	if dubbing == nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Dubbing not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, dubbing)
// }
// func (app *application) getVideoQualities(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
// 		return
// 	}
// 	qualities, err := app.models.Videos.GetQualities(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video qualities"})
// 		return
// 	}
// 	if qualities == nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Video qualities not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, qualities)
// }
