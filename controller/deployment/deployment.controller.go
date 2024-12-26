package deployment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment"
	projectModel "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

// create Deployment
func CreateDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	reqCtx := ctx.Request.Context()

	// get project id - post request
	body, err := getCreateDeploymentBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidBodyMessage)
	}

	// check valid project from id
	project, err := model.GetProjectById(reqCtx, body.ProjectId)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.ProjectNotFoundMessage)
	}

	// check if any deployment is queued
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

	// Spin ECS Task
	_, err = spinEcsTask(reqCtx, deploymentId, project)
	if err != nil {
		logger.Log.Errorln(err)
		_ = tx.Rollback()
		logger.WithRequest(ctx).Panicln("error while launching container")
	}

	// Commit transaction if all steps succeed
	if err := tx.Commit(); err != nil {
		logger.Log.Errorln(err)
		logger.WithRequest(ctx).Panicln("error while committing transaction")
	}

	// Return successful response
	ctx.JSON(http.StatusOK, gin.H{
		"error":  false,
		"status": "queued",
		"data":   gin.H{"deploymentId": deploymentId},
	})
}

// spinEcsTask launches an ECS task
func spinEcsTask(ctx context.Context, deploymentId int, project projectModel.Project) (string, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     constants.AWS_ACCESS_KEY_ID,
				SecretAccessKey: constants.AWS_SECRET_ACCESS_KEY,
				Source:          "CustomEnvironment",
			}, nil
		})),
		config.WithRegion("ap-south-1"),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}

	ecsClient := ecs.NewFromConfig(cfg)
	var taskCount int32 = constants.LaunchTaskCount

	// Define task input parameters
	taskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(constants.ClusterARN),
		TaskDefinition: aws.String(constants.TaskDefinitionARN),
		Count:          &taskCount,
		LaunchType:     types.LaunchTypeFargate,
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name: aws.String(constants.TaskDefinitionContainerName),
					Environment: []types.KeyValuePair{
						{Name: aws.String("ENVIRONMENT"), Value: aws.String(constants.ENV_DEV)},
						{Name: aws.String("APP_NAME"), Value: aws.String(constants.TaskDefinitionENVAppName)},
						{Name: aws.String("BUILD_TEST_URL"), Value: aws.String(constants.TaskDefinitionBuildTestUrl)},
						{Name: aws.String("PROJECT_ID"), Value: aws.String(fmt.Sprint(project.Id))},
						{Name: aws.String("DEPLOYMENT_ID"), Value: aws.String(fmt.Sprint(deploymentId))},
						{Name: aws.String("EMAIl_QUEUE_URL"), Value: aws.String(constants.TaskDefinitionEmailQueueUrl)},
						{Name: aws.String("RECIPIENT_EMAIL"), Value: aws.String(project.UserId)},
						{Name: aws.String("LOG_QUEUE_URL"), Value: aws.String(constants.TaskDefinitionLogQueueUrl)},
						{Name: aws.String("STATUS_QUEUE_URL"), Value: aws.String(constants.TaskDefinitionStatusQueueUrl)},
						{Name: aws.String("S3_BUCKET_NAME"), Value: aws.String(constants.TaskDefinitionS3BucketName)},
						{Name: aws.String("AWS_REGION"), Value: aws.String(constants.AWS_REGION)},
						{Name: aws.String("AWS_ACCESS_KEY_ID"), Value: aws.String(constants.AWS_ACCESS_KEY_ID)},
						{Name: aws.String("AWS_SECRET_ACCESS_KEY"), Value: aws.String(constants.AWS_SECRET_ACCESS_KEY)},
						{Name: aws.String("GIT_REPOSITORY_URL"), Value: aws.String(project.SourceCodeUrl)},
					},
				},
			},
		},
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        []string{constants.TaskDefinitionSubnet1, constants.TaskDefinitionSubnet2, constants.TaskDefinitionSubnet3},
				SecurityGroups: []string{constants.TaskDefinitionSecurityGroup1},
				AssignPublicIp: types.AssignPublicIpEnabled,
			},
		},
	}

	// Run the ECS task
	taskOutput, err := ecsClient.RunTask(ctx, taskInput)
	if err != nil {
		logger.Log.Println("ECS RunTask error: ", err)
		return "", err
	}

	// Extract and return the task ID
	if len(taskOutput.Tasks) > 0 {
		return aws.ToString(taskOutput.Tasks[0].TaskArn), nil
	}

	return "", fmt.Errorf("no tasks were launched: %v", taskOutput.Failures)
}

// get Deployment
func GetDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// get all user Deployment
func GetAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete Deployment - remove from s3
func DeleteDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}

// delete all user Deployment - remove from s3
func DeleteAllDeployment(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
	})
}
