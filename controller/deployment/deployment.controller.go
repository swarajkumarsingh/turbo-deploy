package deployment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment"
)

func CreateDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	body, err := getCreateDeploymentBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidBodyMessage)
	}

	project, err := model.GetProjectById(reqCtx, body.ProjectId)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.ProjectNotFoundMessage)
	}

	exists, err := deploymentAlreadyQueued(reqCtx, project.Id)
	if err != nil {
		logger.WithRequest(ctx).Panicln(messages.SomethingWentWrongMessage)
	}
	if exists {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, "deployment already queued")
	}

	var database = db.Mgr.DBConn
	tx, err := database.BeginTx(reqCtx, nil)
	if err != nil {
		logger.Log.Errorln(err)
		logger.WithRequest(ctx).Panicln("error starting transaction")
	}

	deploymentId, err := model.CreateDeploymentTx(reqCtx, tx, project.Id, project.UserId)
	if err != nil {
		logger.Log.Errorln(err)
		_ = tx.Rollback()
		logger.WithRequest(ctx).Panicln("error while creating deployment")
	}

	_, err = spinEcsTask(reqCtx, deploymentId, project)
	if err != nil {
		logger.Log.Errorln(err)
		_ = tx.Rollback()
		logger.WithRequest(ctx).Panicln("error while launching container")
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Errorln(err)
		logger.WithRequest(ctx).Panicln("error while committing transaction")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":  false,
		"status": "queued",
		"data":   gin.H{"deploymentId": deploymentId},
	})
}

func GetDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	id, valid := getDeploymentIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidDeploymentIdMessage)
	}

	deployment, err := model.GetDeploymentById(reqCtx, id)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.DeploymentNotFoundMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":      false,
		"deployment": deployment,
	})
}

func GetDeploymentStatus(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	deploymentId, valid := getDeploymentIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidProjectIdMessage)
	}

	status, err := model.GetDeploymentStatus(reqCtx, deploymentId)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.ProjectNotFoundMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":  false,
		"status": status,
	})
}

func GetAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "deployment deleted successfully",
	})
}

func DeleteDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})

}

func DeleteAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "all deployment deleted successfully",
	})
}
