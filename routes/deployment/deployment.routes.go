package deploymentRoutes

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/authentication"
	"github.com/swarajkumarsingh/turbo-deploy/controller/deployment"
)

func AddRoutes(router *gin.Engine) {
	r := router.Group("/")

	r.POST("/deployment", deployment.CreateDeployment)
	r.GET("/deployment/:id", deployment.GetDeployment)
	r.GET("/deployment/:id/status", deployment.GetDeploymentStatus)
	r.GET("/deployment", authentication.AuthorizeUser, deployment.GetAllDeployment)
	r.DELETE("/deployment/:id", deployment.DeleteDeployment)
	r.DELETE("/deployment", authentication.AuthorizeUser, deployment.DeleteAllDeployment)
}
