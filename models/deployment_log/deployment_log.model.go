package deploymentlog

import (
	"database/sql"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func GetDeploymentLogsPaginatedValue(deployment_id, itemsPerPage, offset int) (*sql.Rows, error) {
	query := `SELECT id, deployment_id, project_id, data, metadata, data_length, created_at FROM log_events WHERE deployment_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	return database.Query(query, deployment_id, itemsPerPage, offset)
}

func GetTotalDeploymentLogsCount(deployment_id int) int {
	var total int
	query := `SELECT COUNT(*) FROM log_events WHERE deployment_id = $1`
	err := database.QueryRow(query, deployment_id).Scan(&total)
	if err != nil {
		return 0
	}
	return total
}