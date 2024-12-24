package constants

import "os"

var STAGE string = os.Getenv("STAGE")

// Server ENV constants
const (
	ENV_PROD  = "prod"
	ENV_UAT   = "uat"
	ENV_DEV   = "dev"
	ENV_LOCAL = "local"
)

const DefaultRateLimiterPerMinute = 10

const DefaultPerPageSize = 10
const DefaultPageSize = 10
const BcryptHashingCost = 8

const VaultKeySuffix = "-vlt"
const UserIdMiddlewareConstant = "userId"

const DefaultSenderEmailId = "swaraj.singh.wearingo@gmail.com"
const DefaultRecipientEmailId = "swaraj.singh.wearingo@gmail.com"
