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
	config.SetAppVariable()

	// Capture database connection
	config.CreateDBConnection()

	config.CreateElasticConnection()

	// Init fiber app
	app := fiber.New(config.Fiber())

	// Init middleware
	middleware.Fiber(app)
	middleware.Log(app)

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
		port = "3000"
	}

	app.Listen(":" + port)

}
