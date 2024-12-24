package conf

import "os"

var DB_URL = os.Getenv("DB_URL")
var AWS_REGION = os.Getenv("AWS_REGION")
var AWS_SQS_URL = os.Getenv("AWS_SQS_URL")
var AWS_ACCESS_KEY = os.Getenv("AWS_ACCESS_KEY")
var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")

const VisibilityTimeout = 10
const MaxNumberOfMessages = 10
