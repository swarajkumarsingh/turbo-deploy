package conf

import "os"

var AWS_TOKEN = os.Getenv("AWS_TOKEN")
var AWS_SQS_URL = os.Getenv("SQS_URL")
var AWS_REGION = os.Getenv("AWS_REGION")
var AWS_ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")

const VisibilityTimeout = 10
const MaxNumberOfMessages = 10
