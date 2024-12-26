package constants

import "os"

var AWS_REGION string = os.Getenv("AWS_REGION")
var AWS_ACCESS_KEY_ID string = os.Getenv("AWS_ACCESS_KEY")
var AWS_SECRET_ACCESS_KEY string = os.Getenv("AWS_SECRET_ACCESS_KEY")

const LaunchTaskCount = 1

var ClusterARN string = os.Getenv("ClusterARN")
var TaskDefinitionARN string = os.Getenv("TaskDefinitionARN")

const TaskDefinitionBuildTestUrl = ""
const TaskDefinitionContainerName = "builder-image"
const TaskDefinitionENVAppName = "turbo-deploy-build-server"

var TaskDefinitionLogQueueUrl string = os.Getenv("LOGS_SQS_URL")
var TaskDefinitionEmailQueueUrl string = os.Getenv("EMAIl_QUEUE_URL")
var TaskDefinitionStatusQueueUrl string = os.Getenv("STATUS_SQS_URL")

var TaskDefinitionSubnet1 string = os.Getenv("TaskDefinitionSubnet1")
var TaskDefinitionSubnet2 string = os.Getenv("TaskDefinitionSubnet2")
var TaskDefinitionSubnet3 string = os.Getenv("TaskDefinitionSubnet3")
var TaskDefinitionS3BucketName string = os.Getenv("TaskDefinitionS3BucketName")
var TaskDefinitionSecurityGroup1 string = os.Getenv("TaskDefinitionSecurityGroup1")
