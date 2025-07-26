package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (app *application) getVideo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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
}

func (app *application) getAllVideos(c *gin.Context) {
	videos, err := app.models.Videos.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}
	c.JSON(http.StatusOK, videos)
}
func (app *application) uploadVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}
}
func (app *application) deleteVideo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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
func (app *application) getVideoInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	info, err := app.models.Videos.GetInfo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video info"})
		return
	}
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video info not found"})
		return
	}
	c.JSON(http.StatusOK, info)
}
func (app *application) getVideoSubtitles(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	subtitles, err := app.models.Videos.GetSubtitles(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video subtitles"})
		return
	}
	if subtitles == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtitles not found"})
		return
	}
	c.JSON(http.StatusOK, subtitles)
}
func (app *application) getVideoDubbing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	dubbing, err := app.models.Videos.GetDubbing(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video dubbing"})
		return
	}
	if dubbing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dubbing not found"})
		return
	}
	c.JSON(http.StatusOK, dubbing)
}
func (app *application) getVideoQualities(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}
	qualities, err := app.models.Videos.GetQualities(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video qualities"})
		return
	}
	if qualities == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video qualities not found"})
		return
	}
	c.JSON(http.StatusOK, qualities)
}
