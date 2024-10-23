package deploymentRoutes

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/controller/deployment"
)

func AddRoutes(router *gin.Engine) {
	r := router.Group("/")
	
	r.POST("/deployment", deployment.CreateDeployment)
	r.GET("/deployment/:id", deployment.GetDeployment)
	r.GET("/deployment", deployment.GetAllDeployment)
	r.DELETE("/deployment", deployment.DeleteAllDeployment)
	r.DELETE("/deployment/:id", deployment.DeleteDeployment)
}
