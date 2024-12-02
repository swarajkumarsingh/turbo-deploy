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
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidProjectIdMessage)
	}

	project, err := model.GetProjectById(reqCtx, pid)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.ProjectNotFoundMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"project": project,
	})
}

// get all user project
func GetAllProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	page := getCurrentPageValue(ctx)
	itemsPerPage := getItemPerPageValue(ctx)
	offset := getOffsetValue(page, itemsPerPage)

	userId, valid := getUserIdFromReq(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	rows, err := model.GetProjectListPaginatedValue(reqCtx, userId, itemsPerPage, offset)
	if err != nil {
		logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveProjectsMessage)
	}
	defer rows.Close()

	projects := make([]gin.H, 0)

	for rows.Next() {
		var id int
		var name, subdomain, language string
		if err := rows.Scan(&id, &name, &subdomain, &language); err != nil {
			logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveProjectsMessage)
		}
		projects = append(projects, gin.H{"id": id, "name": name, "subdomain": subdomain, "language": language})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users":       projects,
		"page":        page,
		"per_page":    itemsPerPage,
		"total_pages": calculateTotalPages(page, itemsPerPage),
	})
}

// update project - projectName, customDomain
func UpdateProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	pid, valid := getProjectIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidProjectIdMessage)
	}

	// get projectName and customDomain
	body, err := getUpdateProjectBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, err)
	}

	// Check sub-domain availability
	available, err := model.IsSubDomainAvailable(reqCtx, body.Subdomain)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusInternalServerError, err)
	}
	if !available {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.SubDomainAlreadyExists)
	}

	// update in project DB
	subDomainAlreadyExists, err := model.UpdateProject(reqCtx, pid, body.Name, body.Subdomain)
	if subDomainAlreadyExists {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, err)
	}
	if err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "project updated successfully",
	})
}

// delete all user project
func DeleteAllProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()
	
	// get project id
	uid, valid := getUserIdFromReq(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	// delete project
	if err := model.DeleteAllProjectFromUser(reqCtx, uid); err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "all projects deleted successfully",
	})
}

// delete project
func DeleteProject(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	// get project id
	pid, valid := getProjectIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	// delete project
	if err := model.DeleteProjectFromUser(reqCtx, pid); err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "project deleted successfully",
	})
}
