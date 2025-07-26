package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	g := gin.Default()
	v1 := g.Group("/api/")

	v1.GET("/", app.getAllVideos)
	v1.POST("/", app.uploadVideo)
	v1.GET("/:id", app.getVideo)
	v1.DELETE("/:id", app.deleteVideo)
	v1.GET("/:id/info", app.getVideoInfo)
	v1.GET("/:id/subtitles", app.getVideoSubtitles)
	v1.GET("/:id/dubbing", app.getVideoDubbing)
	v1.GET("/:id/qualities", app.getVideoQualities)
	return g
}
