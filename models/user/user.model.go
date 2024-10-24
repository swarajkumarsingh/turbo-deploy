package userModel

import (
	"context"
	"errors"

	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	"github.com/swarajkumarsingh/turbo-deploy/infra/db"
)

var database = db.Mgr.DBConn

func UserAlreadyExistsWithUsername(username string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)"

	err := database.QueryRow(query, username).Scan(&exists)

	if err != nil {
		logger.Log.Println(err)
		return false
	}

	return exists
}

func InsertUser(body UserBody, password string) error {
	query := `INSERT INTO users(username, firstname, lastname, email, password, phone, address) VALUES($1, $2, $3, $4, $5, $6, $7)`
	_, err := database.Exec(query, body.Username, body.FirstName, body.LastName, body.Email, password, body.Phone, body.Address)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByUsername(context context.Context, username string) (User, error) {
	var userModel User
	validUserName := general.ValidUserName(username)
	if !validUserName {
		return userModel, errors.New("invalid username")
	}

	query := "SELECT * FROM users WHERE username = $1"
	err := database.GetContext(context, &userModel, query, username)
	if err == nil {
		return userModel, nil
	}
	return userModel, err
}

func GetUserIdByUsername(context context.Context, username string) (int, error) {
    var userId int
    validUserName := general.ValidUserName(username)
    if !validUserName {
        return 0, errors.New("invalid username")
    }

    query := "SELECT id FROM users WHERE username = $1"
    err := database.GetContext(context, &userId, query, username)
    if err == nil {
        return userId, nil
    }
    return 0, err
}
