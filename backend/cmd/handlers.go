package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ksamf/video-upscaling/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
)

func (app *application) uploadVideo(c *gin.Context) {
	conf := config.New()
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
	fileName := uuid.New()

	savePathS3 := "https://" + conf.S3.EndpointURL + "/" + conf.S3.BucketName + "/" + fileName.String()
	err = app.models.Videos.Insert(&database.Video{
		VideoId:   fileName,
		Name:      strings.TrimSuffix(header.Filename, ext),
		VideoPath: savePathS3,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка сохранения файла: %v", err)
	}

	if err = utils.Handler(file, fileName, app.models.Videos); err != nil {
		c.String(http.StatusInternalServerError, "Failed transcode video: %v", err)
	}

	c.String(http.StatusOK, fmt.Sprintf("Видео успешно загружено: %s", savePathS3))
}

func (app *application) getVideo(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
	}
	video, err := app.models.Videos.Get(id)
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
	videos, err := app.models.Videos.GetAll()
	if err != nil {
		fmt.Println(err)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
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
	fmt.Println(updateVideo)
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

// func (app *application) getVideoInfo(c *gin.Context) {
// 	id, err := strconv.Atoi(c.Param("id"))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
// 		return
// 	}
// 	info, err := app.models.Videos.GetInfo(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video info"})
// 		return
// 	}
// 	if info == nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Video info not found"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, info)
// }
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
