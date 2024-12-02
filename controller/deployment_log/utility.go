package deployment_logs

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
)

func getDeploymentIdFromReq(ctx *gin.Context) (int, bool) {
	deploymentId := ctx.Param("id")
	valid := general.SQLInjectionValidation(deploymentId)

	if !valid {
		return 0, false
	}
	id, err := general.IsInt(deploymentId)
	if err != nil {
		return 0, false
	}

	return id, true
}

func getCurrentPageValue(ctx *gin.Context) int {
	pageStr := ctx.DefaultQuery("page", strconv.Itoa(constants.DefaultPageSize))
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		logger.WithRequest(ctx).Errorln("Invalid page value; defaulting to 1:", err)
		return 1
	}
	return page
}

func getItemPerPageValue(ctx *gin.Context) int {
	perPageStr := ctx.DefaultQuery("per_page", strconv.Itoa(constants.DefaultPerPageSize))
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 || perPage > 100 {
		logger.WithRequest(ctx).Errorln("Invalid per_page value; defaulting to:", constants.DefaultPerPageSize)
		return constants.DefaultPerPageSize
	}
	return perPage
}

func getOffsetValue(page int, itemsPerPage int) int {
	if page < 1 {
		page = 1
	}
	if itemsPerPage < 1 {
		itemsPerPage = constants.DefaultPerPageSize
	}
	return (page - 1) * itemsPerPage
}


func calculateTotalPages(totalLogs, itemsPerPage int) int {
	if itemsPerPage <= 0 {
		return 1
	}
	totalPages := (totalLogs + itemsPerPage - 1) / itemsPerPage
	return totalPages
}
