package deploymentlog

import (
	"database/sql"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func GetDeploymentLogsPaginatedValue(deployment_id, itemsPerPage, offset int) (*sql.Rows, error) {
	query := `SELECT id, deployment_id, project_id, data, metadata, data_length, created_at FROM projects WHERE deployment_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	return database.Query(query, deployment_id, itemsPerPage, offset)
}