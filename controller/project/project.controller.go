package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	model "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

func CreateProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	// Get request body
	body, err := getCreateProjectBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidBodyMessage)
	}

	// Check sub-domain availability
	available, err := model.IsSubDomainAvailable(reqCtx, body.Subdomain)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusInternalServerError, err)
	}
	if !available {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.SubDomainAlreadyExists)
	}

	// Validate GitHub URL format
	if !IsValidGitHubURL(body.SourceCodeUrl) {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidSourceURLMessage)
	}

	// Asynchronous github repo validation
	resultChan := make(chan bool)
	go ValidateGitHubURL(body.SourceCodeUrl, resultChan)
	valid := <-resultChan
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.GithubRepoNotFoundOrPrivate)
	}

	// Add to project table
	subDomainAlreadyExists, err := model.CreateProject(reqCtx, body)
	if subDomainAlreadyExists {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, err)
	}
	if err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "project created successfully",
	})
}

// get project
func GetProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	pid, valid := getProjectIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	user, err := model.GetProjectById(reqCtx, pid)
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
