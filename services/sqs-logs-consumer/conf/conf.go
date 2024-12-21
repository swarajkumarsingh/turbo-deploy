package conf

import "os"

var AWS_TOKEN = os.Getenv("AWS_TOKEN")
var AWS_REGION = ""
var AWS_SQS_URL = ""
var AWS_ACCESS_KEY = ""
var AWS_SECRET_ACCESS_KEY = ""
var DB_URL=""


const VisibilityTimeout = 10
const MaxNumberOfMessages = 10