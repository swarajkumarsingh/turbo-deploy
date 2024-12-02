package deployment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment"
)

// create Deployment
func CreateDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	// get project id - post request
	body, err := getCreateDeploymentBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidBodyMessage)
	}

	// check valid project from id
	project, err := model.GetProjectById(reqCtx, body.ProjectId)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.ProjectNotFoundMessage)
	}

	// check if any deployment is queued
	exists, err := deploymentAlreadyQueued(reqCtx, project.Id)
	if err != nil {
		logger.WithRequest(ctx).Panicln(messages.SomethingWentWrongMessage)
	}
	if exists {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, "deployment already queued")
	}

	// TODO: spin the container & when build finishes update deployment table fields

	// create deployment & return id
	if err = model.CreateDeployment(reqCtx, project.Id, project.UserId); err != nil {
		logger.WithRequest(ctx).Panicln(messages.SomethingWentWrongMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":  false,
		"status": "queued",
		"data":   gin.H{"deploymentId": "id"},
	})
}

// get Deployment
func GetDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// get all user Deployment
func GetAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete Deployment - remove from s3
func DeleteDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete all user Deployment - remove from s3
func DeleteAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}
