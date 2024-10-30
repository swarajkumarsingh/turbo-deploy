package project

import (
	"context"

	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func GetProjectById(context context.Context, pid int) (Project, error) {
	var model Project
	query := "SELECT * FROM users WHERE id = $1"
	err := database.GetContext(context, &model, query, pid)
	if err == nil {
		return model, nil
	}
	return model, err
}
