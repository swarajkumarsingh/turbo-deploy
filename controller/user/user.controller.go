package user

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/errorHandler"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	model "github.com/swarajkumarsingh/turbo-deploy/models/user"
)

// create user
func CreateUser(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	body, err := getCreateUserBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, err)
	}

	if model.UserAlreadyExistsWithUsername(body.Username) {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.UserAlreadyExistsMessageWithUsername)
	}

	if model.UserAlreadyExistsWithEmail(body.Email) {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.UserAlreadyExistsMessageWithEmail)
	}

	hashedPassword, err := hashPassword(body.Password)
	if err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	if err = model.InsertUser(body, hashedPassword); err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	id, err := model.GetUserIdByUsername(context.TODO(), body.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.UserNotFoundMessage)
		}
		logger.WithRequest(ctx).Panicln(err)
	}

	token, err := generateJwtToken(strconv.Itoa(id))
	if err != nil {
		logger.WithRequest(ctx).Panicln("unable to login, try again later")
	}

	if err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "User Created successfully",
		"token":   token,
	})
}

// get user
func GetUser(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)
	uid, valid := getUserIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	user, err := model.GetUserById(context.TODO(), uid)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusNotFound, messages.UserNotFoundMessage)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
		"user":  user,
	})
}

// update user
func UpdateUser(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	uid, valid := getUserIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	body, err := getUpdateUserBody(ctx)
	if err != nil {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidBodyMessage)
	}

	if err = model.UpdateUser(context.TODO(), uid, body); err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User updated successfully",
	})
}

// delete user
func DeleteUser(ctx *gin.Context) {
	defer errorHandler.Recovery(ctx, http.StatusConflict)

	uid, valid := getUserIdFromParam(ctx)
	if !valid {
		logger.WithRequest(ctx).Panicln(http.StatusBadRequest, messages.InvalidUserIdMessage)
	}

	if err := model.DeleteUser(context.TODO(), uid); err != nil {
		logger.WithRequest(ctx).Panicln(err)
	}

	// TODO: Delete all projects with user id

	ctx.JSON(http.StatusOK, gin.H{
		"error": false,
		"message": "User deleted successfully",
	})
}
