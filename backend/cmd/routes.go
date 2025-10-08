package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	router := gin.Default()
	router.POST("/upload", app.uploadVideo)
	router.GET("/video", app.getAllVideos)
	router.GET("/video/:id", app.getVideo)
	router.PATCH("/video/:id", app.updateVideoPartial)
	router.DELETE("/video/:id", app.deleteVideo)
	router.GET("/video/:id/sub", app.getVideoSubtitles)
	// router.GET("/video/:id/dub", app.getVideoDubbing)
	return router

}
