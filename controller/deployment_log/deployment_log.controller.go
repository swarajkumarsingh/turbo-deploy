package deployment_logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment_log"
)

// get all user Deployment - from logs table
func GetDeploymentLogs(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	page := getCurrentPageValue(ctx)
	itemsPerPage := getItemPerPageValue(ctx)
	offset := getOffsetValue(page, itemsPerPage)

	deploymentId, valid := getDeploymentIdFromReq(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	rows, err := model.GetDeploymentLogsPaginatedValue(deploymentId, itemsPerPage, offset)
	if err != nil {
		logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveDeploymentLogsMessage)
	}
	defer rows.Close()

	logs := make([]gin.H, 0)

	for rows.Next() {
		var id int
		var deployment_id, project_id, data, metadata, data_length, created_at string
		if err := rows.Scan(&id, &deployment_id, &project_id, &data, &metadata, &data_length, &created_at); err != nil {
			logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveDeploymentLogsMessage)
		}
		logs = append(logs, gin.H{"id": id, "deployment_id": deployment_id, "project_id": project_id, "data": data, "metadata": metadata, "data_length": data_length, "created_at": created_at})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"page":        page,
		"per_page":    itemsPerPage,
		"total_pages": calculateTotalPages(page, itemsPerPage),
	})
}
