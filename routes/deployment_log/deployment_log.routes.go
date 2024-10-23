package deploymentLogRoutes

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/controller/deployment_log"
)

func AddRoutes(router *gin.Engine) {
	r := router.Group("/")
	
	r.GET("/deployment/:id/logs", deployment_logs.GetDeploymentLogs)
}
