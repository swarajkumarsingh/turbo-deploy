package userRoutes

import (
	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/controller/user"
)

func AddRoutes(router *gin.Engine) {
	r := router.Group("/")
	
	r.POST("/user", user.CreateUser)
	r.GET("/user/:uid", user.GetUser)
	r.PATCH("/user/:uid", user.UpdateUser)
	r.DELETE("/user/:uid", user.DeleteUser)
}
