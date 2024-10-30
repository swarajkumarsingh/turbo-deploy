package project

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
)

func getProjectIdFromParam(ctx *gin.Context) (int, bool) {
	userId := ctx.Param("uid")
	valid := general.SQLInjectionValidation(userId)

	if !valid {
		return 0, false
	}
	uid, err := general.IsInt(userId)
	if err != nil {
		return 0, false
	}

	return uid, true
}
