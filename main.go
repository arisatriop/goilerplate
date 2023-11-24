package main

import (
	"fmt"
	"os"

	"goilerplate/config"
	"goilerplate/delivery/http/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set up Application Variable
	config.SetAppVariable()

	// Capture database connection
	config.CreateDBConnection()

	// Run server
	router := gin.Default()
	routes.InitRoute(router)

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

	router.Run(":" + port)
}
