package userModel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
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

func UserAlreadyExistsWithEmail(email string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)"

	err := database.QueryRow(query, email).Scan(&exists)

	if err != nil {
		logger.Log.Println(err)
		return false
	}

	return exists
}

func InsertUser(body UserBody, password string) error {
	query := `INSERT INTO users(username, firstname, lastname, email, password, phone, address) VALUES($1, $2, $3, $4, $5, $6, $7)`
	data, err := database.Exec(query, body.Username, body.FirstName, body.LastName, body.Email, password, body.Phone, body.Address)
	fmt.Println("Hello ", data)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByUsernameWithUserId(context context.Context, userId string) (User, error) {
	var userModel User
	validUserName := general.SQLInjectionValidation(userId)
	if !validUserName {
		return userModel, errors.New("invalid username")
	}

	query := "SELECT * FROM users WHERE id = $1"
	err := database.GetContext(context, &userModel, query, userId)
	if err == nil {
		return userModel, nil
	}
	return userModel, err
}

func CheckIfUsernameExistsWithId(context context.Context, userId string) (User, error) {
	var user User
	user, err := GetUserByUsernameWithUserId(context, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New(messages.UserNotFoundMessage)
		}
		return user, err
	}

	return user, nil
}

func UpdateUser(context context.Context, uid int, body UserUpdateBody) error {
	query := "UPDATE users SET username = $2, firstname = $3, lastname = $4, address = $5, experience = $6, primary_goal = $7, user_role = $8, plan_type = $9 WHERE id = $1"
	res, err := database.ExecContext(context, query, uid, body.Username, body.FirstName, body.LastName, body.Address, body.Experience, body.PrimaryGoal, body.UserRole, body.PlanType)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.New("could not update user")
	}

	return nil
}

func DeleteUser(ctx context.Context, uid int) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := database.ExecContext(ctx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("user not found or user already deleted")
	}

	return nil
}

func DeleteProjectFromUser(ctx context.Context, uid int) error {
	query := "DELETE from projects WHERE user_id = $1"

	result, err := database.ExecContext(ctx, query, uid)
	if err != nil {
		return errors.New("user not found or already deleted")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}

	return nil
}

func GetUserByUsername(context context.Context, username string) (User, error) {
	var userModel User
	validUserName := general.SQLInjectionValidation(username)
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

func GetUserById(context context.Context, uid int) (User, error) {
	var userModel User
	query := "SELECT * FROM users WHERE id = $1"
	err := database.GetContext(context, &userModel, query, uid)
	if err == nil {
		return userModel, nil
	}
	return userModel, err
}

func GetUserIdByUsername(context context.Context, username string) (int, error) {
	var userId int
	validUserName := general.SQLInjectionValidation(username)
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
