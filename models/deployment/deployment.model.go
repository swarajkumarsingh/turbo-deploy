package deployment

import (
	"context"
	"database/sql"

	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
	projectModel "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

var database = db.Mgr.DBConn

func GetProjectById(context context.Context, pid string) (projectModel.Project, error) {
	var model projectModel.Project
	query := "SELECT * FROM projects WHERE id = $1"
	err := database.GetContext(context, &model, query, pid)
	if err == nil {
		return model, nil
	}
	return model, err
}

func GetQueuedProjectCount(context context.Context, project_id int) (int, error) {
	var total int
	query := `SELECT COUNT(*) FROM deployments WHERE project_id = $1 AND status = $2`
	err := database.QueryRowContext(context, query, project_id, "QUEUE").Scan(&total)
	if err != nil {
		return total, err
	}
	return total, nil
}

func CreateDeployment(ctx context.Context, projectId int, userId string) (int, error) {
	var deploymentId int
	query := `INSERT INTO deployments(user_id, project_id) VALUES($1, $2) RETURNING id`
	err := database.QueryRowContext(ctx, query, userId, projectId).Scan(&deploymentId)
	if err != nil {
		return 0, err
	}
	return deploymentId, nil
}

func CreateDeploymentTx(ctx context.Context, tx *sql.Tx, projectId int, userId string) (int, error) {
	var deploymentId int
	query := `INSERT INTO deployments(user_id, project_id) VALUES($1, $2) RETURNING id`
	err := tx.QueryRowContext(ctx, query, userId, projectId).Scan(&deploymentId)
	if err != nil {
		return 0, err
	}
	return deploymentId, nil
}
