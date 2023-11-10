package routes

import (
	"github.com/arisatriop/goilerplate/config"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RouteGroup struct {
	Public *gin.RouterGroup
}

func InitRoute(gin *gin.Engine) {
	db := config.GetDBConnection()
	validator := validator.New()

	_ = db
	_ = validator
}
