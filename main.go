package main

import (
	"fmt"
	"goilerplate/api/middleware"
	"goilerplate/api/route"
	"goilerplate/config"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {

	// Set up Application Variable
	appVariable := config.SetAppVariable()
	appVariable.DB = config.CreateDBConnection()

	if appVariable.CacheDriver == "redis" {
		appVariable.RedisClient = config.CreateRedisConnection()
	}

	if appVariable.LogChannel == "elasticsearch" {
		appVariable.ElasticClient = config.CreateElasticConnection()
	}

	// Init fiber app
	app := fiber.New(config.Fiber())

	// Init middleware
	middleware.Fiber(app)
	middleware.Log(app)
	middleware.Recover(app)
	middleware.Auth(app)

	// Init route
	route.Init(app)

	fmt.Println("")
	fmt.Println("")
	fmt.Println("=========================================================================================")
	fmt.Println("===================================== READY TO SERVE ====================================")
	fmt.Println("=========================================================================================")
	fmt.Println("")
	fmt.Println("")

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	app.Listen(":" + port)

}
