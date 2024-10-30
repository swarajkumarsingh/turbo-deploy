package project

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	model "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

// create project
func CreateProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// get project
func GetProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	pid, valid := getProjectIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	user, err := model.GetProjectById(context.TODO(), pid)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.UserNotFoundMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
		"user":  user,
	})
}

// get all user project
func GetAllProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// update project - projectName, customDomain
func UpdateProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete project
func DeleteProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete all user project
func DeleteAllProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}
