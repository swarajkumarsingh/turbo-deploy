package db

import (
	"os"

	"github.com/swarajkumarsingh/turbo-deploy/constants"
)

const (
	ENV_PROD  = constants.ENV_PROD
	ENV_UAT   = constants.ENV_UAT
	ENV_DEV   = constants.ENV_DEV
	ENV_LOCAL = constants.ENV_LOCAL
)

var DB_URL string = os.Getenv("DB_URL")
