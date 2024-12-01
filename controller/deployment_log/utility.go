package deployment_logs

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
)

func getDeploymentIdFromReq(ctx *gin.Context) (int, bool) {
	deploymentId := ctx.Param("deploymentId")
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
	val, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		logger.WithRequest(ctx).Errorln("error while extracting current page value: ", err)
		return 1
	}
	return val
}

func getOffsetValue(page int, itemsPerPage int) int {
	return (page - 1) * itemsPerPage
}

func getItemPerPageValue(ctx *gin.Context) int {
	val, err := strconv.Atoi(ctx.DefaultQuery("per_page", strconv.Itoa(constants.DefaultPerPageSize)))
	if err != nil {
		logger.WithRequest(ctx).Errorln("error while extracting item per-page value: ", err)
		return constants.DefaultPerPageSize
	}
	return val
}

func calculateTotalPages(page, itemsPerPage int) int {
	if page <= 0 {
		return 1
	}
	return (page + itemsPerPage - 1) / itemsPerPage
}
