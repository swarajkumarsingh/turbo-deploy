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
	reqCtx := ctx.Request.Context()

	page := getCurrentPageValue(ctx)
	itemsPerPage := getItemPerPageValue(ctx)
	offset := getOffsetValue(page, itemsPerPage)

	deploymentId, valid := getDeploymentIdFromReq(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidDeploymentIdMessage)
	}

	rows, err := model.GetDeploymentLogsPaginatedValue(reqCtx, deploymentId, itemsPerPage, offset)
	if err != nil {
		logger.WithRequest(ctx).Errorln(deploymentId, itemsPerPage, offset, err.Error())
		logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveDeploymentLogsMessage)
	}
	defer rows.Close()

	logs := make([]gin.H, 0)

	for rows.Next() {
		var id int
		var logEntry gin.H
		var deployment_id, project_id, message, stack, log_type, timestamp string

		if err := rows.Scan(&id, &deployment_id, &project_id, &message, &stack, &log_type, &timestamp); err != nil {
			logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveDeploymentLogsMessage)
		}

		logEntry = gin.H{
			"id":            id,
			"deployment_id": deployment_id,
			"project_id":    project_id,
			"message":          message,
			"stack":      stack,
			"log_type":   log_type,
			"created_at":    timestamp,
		}
		logs = append(logs, logEntry)
	}

	if err := rows.Err(); err != nil {
		logger.WithRequest(ctx).Panicln(messages.FailedToRetrieveDeploymentLogsMessage)
		return
	}

	totalLogs := model.GetTotalDeploymentLogsCount(reqCtx, deploymentId)

	ctx.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"page":        page,
		"per_page":    itemsPerPage,
		"total_logs":  totalLogs,
		"total_pages": calculateTotalPages(totalLogs, itemsPerPage),
	})
}
