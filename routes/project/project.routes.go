package projectRoutes

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/authentication"
	"github.com/swarajkumarsingh/turbo-deploy/controller/project"
)

func AddRoutes(router *gin.Engine) {
	r := router.Group("/")
	
	r.POST("/project", project.CreateProject)
	r.GET("/project/:pid", project.GetProject)
	r.GET("/projects", authentication.AuthorizeUser, project.GetAllProject)
	r.PATCH("/project/:pid", project.UpdateProject)
	r.DELETE("/project/:pid", project.DeleteProject)
}
