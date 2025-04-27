package routes

import (
	"github.com/Tsarbomba69-com/mammoth.server/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterProjectRoutes(http *gin.Engine) {
	r := http.Group("/api/v1/projects")
	{
		r.POST("/", controllers.CreateProject)
		r.GET("/", controllers.GetProjects)
		r.GET("/:id/compare", controllers.Compare)
		r.GET("/:id/dump", controllers.Dump)
	}
}
