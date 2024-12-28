package deployment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func GetDeploymentById(context context.Context, id int) (Deployment, error) {
	var model Deployment
	query := "SELECT * FROM deployments WHERE id = $1"
	err := database.GetContext(context, &model, query, id)
	if err == nil {
		return model, nil
	}
	return model, err
}

func GetDeploymentStatus(context context.Context, id int) (string, error) {
	var status string
	query := "SELECT status FROM deployments WHERE id = $1"
	err := database.GetContext(context, &status, query, id)
	if err == nil {
		return status, nil
	}
	return status, err
}

func GetDeploymentListPaginatedValue(context context.Context, uid string, itemsPerPage, offset int) (*sql.Rows, error) {
	query := `SELECT id, project_id, status, ready_url FROM deployments WHERE user_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	return database.QueryContext(context, query, uid, itemsPerPage, offset)
}

func DeleteDeploymentFromUser(ctx context.Context, pid int) error {
	query := "DELETE from deployments WHERE id = $1"

	result, err := database.ExecContext(ctx, query, pid)
	if err != nil {
		return errors.New("deployment not found or already deleted")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("deployment not found or already deleted")
	}

	return nil
}

func GetAllDeploymentIDsByUser(ctx context.Context, uid string) ([]int, error) {
	query := "SELECT id FROM deployments WHERE user_id = $1"

	rows, err := database.QueryContext(ctx, query, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment IDs: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan deployment ID: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return ids, nil
}

func DeleteAllDeploymentFromUser(ctx context.Context, uid string) error {
	query := "DELETE from deployments WHERE user_id = $1"

	result, err := database.ExecContext(ctx, query, uid)
	if err != nil {
		return errors.New("deployment not found or already deleted")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("deployment not found or already deleted")
	}

	return nil
}