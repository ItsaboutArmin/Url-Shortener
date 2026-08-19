package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *URLHandler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.POST("/shorten", handler.Shorten)
	}

	r.GET("/:code", handler.Redirect)

	r.Static("/static", "./web")
	r.StaticFile("/", "./web/index.html")

	return r
}
