package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func GetProjectById(context context.Context, pid int) (Project, error) {
	var model Project
	query := "SELECT * FROM projects WHERE id = $1"
	err := database.GetContext(context, &model, query, pid)
	if err == nil {
		return model, nil
	}
	return model, err
}

func GetProjectListPaginatedValue(context context.Context, uid string, itemsPerPage, offset int) (*sql.Rows, error) {
	query := `SELECT id, name, subdomain, language FROM projects WHERE user_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	return database.QueryContext(context, query, uid, itemsPerPage, offset)
}

func IsSubDomainAvailable(ctx context.Context, subDomain string) (bool, error) {
	query := `SELECT COUNT(*) FROM projects WHERE subdomain = $1`
	var count int
	err := database.QueryRowContext(ctx, query, subDomain).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func DeleteAllProjectFromUser(ctx context.Context, uid string) error {
	query := "DELETE from projects WHERE user_id = $1"

	result, err := database.ExecContext(ctx, query, uid)
	if err != nil {
		return errors.New("projects not found or already deleted")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("projects not found or already deleted")
	}

	return nil
}

func DeleteProjectFromUser(ctx context.Context, pid int) error {
	query := "DELETE from projects WHERE id = $1"

	result, err := database.ExecContext(ctx, query, pid)
	if err != nil {
		return errors.New("project not found or already deleted")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("project not found or already deleted")
	}

	return nil
}

func CreateProject(context context.Context, body ProjectBody) (bool, error) {
	query := `INSERT INTO projects(user_id, name, source_code_url, subdomain, custom_domain, source_code, language, is_dockerized) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := database.ExecContext(context, query, body.UserId, body.Name, body.SourceCodeUrl, body.Subdomain, body.Subdomain, body.SourceCode, body.Language, body.IsDockerized)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return true, errors.New(messages.SubDomainAlreadyExists)
		}
		return false, err
	}
	return false, nil
}

func UpdateProject(ctx context.Context, id int, name, subDomain string) (bool, error) {
	query := `UPDATE projects SET name = $1, subdomain = $2 WHERE id = $3;`
	_, err := database.ExecContext(ctx, query, name, subDomain, id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return true, errors.New(messages.SubDomainAlreadyExists)
		}
		return false, err
	}
	return false, nil
}
