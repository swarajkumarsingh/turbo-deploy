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

const DefaultSenderEmailId = "swaraj.singh.wearingo@gmail.com"
const DefaultRecipientEmailId = "swaraj.singh.wearingo@gmail.com"

const OtpAESEncryptKey = "fjnsfjsdnfjs"

const StatusBanSeller = "ban"
const StatusActiveSeller = "active"
const StatusSuspendSeller = "suspend"

const UserIdMiddlewareConstant = "userId"
const SellerIdMiddlewareConstant = "sellerId"
