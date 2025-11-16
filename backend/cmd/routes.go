package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) routes() http.Handler {
	router := gin.Default()
	gin.SetMode(app.config.App.Debug)

	router.POST("/signup", app.auth.Signup)
	router.POST("/login", app.auth.Login)
	router.POST("/logout", app.auth.AuthMiddleware, app.auth.Logout)

	router.GET("/video", app.auth.OptionalAuthMiddleware, app.handlers.Video.GetAll) //query limit=10, offset=0
	router.GET("/video/:id", app.auth.OptionalAuthMiddleware, app.handlers.Video.Get)
	router.POST("/video", app.auth.AuthMiddleware, app.handlers.Video.Upload) //query name=video1
	router.DELETE("/video/:id", app.auth.AuthMiddleware, app.handlers.Video.Delete)
	router.POST("/video/:id", app.auth.AuthMiddleware, app.handlers.Action.Reaction) //query reaction=like||dislike
	router.POST("/video/:id/upscale", app.handlers.Video.Upscale)                    //query real=true||false
	//default create sub for original language and english for translate sub query lang=fr||ru||...
	router.GET("/video/:id/sub", app.handlers.Video.Subtitles)

	router.GET("/user/:id", app.handlers.User.Get)
	router.DELETE("/user/:id", app.handlers.User.Delete)
	router.GET("/user/:id/video", app.auth.AuthMiddleware, app.handlers.User.Videos)
	return router
}
