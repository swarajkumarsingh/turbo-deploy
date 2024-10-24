package user

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/swarajkumarsingh/turbo-deploy/conf"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	validators "github.com/swarajkumarsingh/turbo-deploy/functions/validator"
	model "github.com/swarajkumarsingh/turbo-deploy/models/user"
	"golang.org/x/crypto/bcrypt"
)

func getCreateUserBody(ctx *gin.Context) (model.UserBody, error) {
	var body model.UserBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return body, errors.New(messages.InvalidBodyMessage)
	}

	if err := validators.ValidateStruct(body); err != nil {
		return body, err
	}
	return body, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), constants.BcryptHashingCost)
	return string(bytes), err
}

func generateJwtToken(userId string) (string, error) {
	expirationTime := time.Now().Add(5 * 24 * time.Hour)
	claims := &model.Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(conf.JWTSecretKey)
	if err != nil {
		return "", err
	}

	if tokenString == "" {
		return "", errors.New("error while authorizing")
	}

	return tokenString, nil
}
