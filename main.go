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

	// Init fiber app
	fiber := fiber.New(config.Fiber())

	// Init middleware
	middleware.Fiber(fiber)
	middleware.Log(fiber)
	middleware.Recover(fiber)

	// Init route
	route.Init(fiber)

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

	fiber.Listen(":" + port)

}
