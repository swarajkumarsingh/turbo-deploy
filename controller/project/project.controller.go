package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
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

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
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
