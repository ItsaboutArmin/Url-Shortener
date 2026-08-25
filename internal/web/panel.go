package web

import (
	"github.com/gin-gonic/gin"
)

func SetupWebRoutes(r *gin.Engine) {

	r.Static("/static", "./web/static")

	r.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	r.GET("/panel", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})
}
