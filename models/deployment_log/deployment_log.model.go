package deploymentlog

import (
	"context"
	"database/sql"

	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func GetDeploymentLogsPaginatedValue(context context.Context, deployment_id, itemsPerPage, offset int) (*sql.Rows, error) {
	query := `SELECT id, deployment_id, project_id, message, stack, log_type, timestamp FROM deployment_logs WHERE deployment_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	return database.QueryContext(context, query, deployment_id, itemsPerPage, offset)
}

func GetTotalDeploymentLogsCount(context context.Context, deployment_id int) int {
	var total int
	query := `SELECT COUNT(*) FROM deployment_logs WHERE deployment_id = $1`
	err := database.QueryRow(query, deployment_id).Scan(&total)
	if err != nil {
		return 0
	}
	return total
}