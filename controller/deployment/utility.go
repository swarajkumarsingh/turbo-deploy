package deployment

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	validators "github.com/swarajkumarsingh/turbo-deploy/functions/validator"
	model "github.com/swarajkumarsingh/turbo-deploy/models/deployment"
	projectModel "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

func getS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(constants.AWS_REGION),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(constants.AWS_ACCESS_KEY_ID, constants.AWS_SECRET_ACCESS_KEY, "")),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS configuration: %v", err))
	}
	return s3.NewFromConfig(cfg)
}

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

func getDeploymentIdFromParam(ctx *gin.Context) (int, bool) {
	userId := ctx.Param("id")
	valid := general.SQLInjectionValidation(userId)

	if !valid {
		return 0, false
	}
	uid, err := general.IsInt(userId)
	if err != nil {
		return 0, false
	}

	return uid, true
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
		Cluster:        aws.String("arn:aws:ecs:ap-south-1:491085393011:cluster/build-cluster"),
		TaskDefinition: aws.String("arn:aws:ecs:ap-south-1:491085393011:task-definition/builder-task:3"),
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
						{Name: aws.String("EMAIl_QUEUE_URL"), Value: aws.String("https://sqs.ap-south-1.amazonaws.com/491085393011/turbo-deploy-email-queue")},
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
				Subnets:        []string{"subnet-0059bc11ef7d6770a", "subnet-0ad58add7254176b2", "subnet-005823699c49d0a22"},
				SecurityGroups: []string{"sg-0e60eba3c7543f186"},
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

func getUserIdFromReq(ctx *gin.Context) (string, bool) {
	uid, valid := ctx.Get(constants.UserIdMiddlewareConstant)
	if !valid || uid == nil || fmt.Sprintf("%v", uid) == "" {
		return "", false
	}

	return fmt.Sprintf("%v", uid), true
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
func deleteS3FilesForDeployment(deploymentID int) error {
	s3Client := getS3Client()
	bucketName := constants.TaskDefinitionS3BucketName
	prefix := fmt.Sprintf("__outputs/%d/", deploymentID)

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}
	listOutput, err := s3Client.ListObjectsV2(context.TODO(), listInput)
	if err != nil {
		return fmt.Errorf("failed to list S3 objects for deployment %d: %w", deploymentID, err)
	}

	// Collect object keys to delete
	var objectsToDelete []s3Types.ObjectIdentifier
	for _, obj := range listOutput.Contents {
		objectsToDelete = append(objectsToDelete, s3Types.ObjectIdentifier{
			Key: obj.Key,
		})
	}

	// Delete the objects
	if len(objectsToDelete) > 0 {
		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3Types.Delete{
				Objects: objectsToDelete,
				Quiet:   aws.Bool(true),
			},
		}
		_, err := s3Client.DeleteObjects(context.TODO(), deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete S3 objects for deployment %d: %w", deploymentID, err)
		}
	}

	return nil
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
