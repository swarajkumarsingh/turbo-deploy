package deployment

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	validators "github.com/swarajkumarsingh/turbo-deploy/functions/validator"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment"
)

func getCreateDeploymentBody(ctx *gin.Context) (model.DeploymentBody, error) {
	var body model.DeploymentBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return body, errors.New(messages.InvalidBodyMessage)
	}

	if err := validators.ValidateStruct(body); err != nil {
		return body, err
	}

	return body, nil
}

func deploymentAlreadyQueued(context context.Context, projectId int) (bool, error) {
	count, err := model.GetQueuedProjectCount(context, projectId)
	if err != nil || count > 0 {
		return true, err
	}
	return false, nil
}
