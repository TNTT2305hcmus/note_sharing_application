package main

import (
	"fmt"
	"log"
	"note_sharing_application/server/routers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Server is booting...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error: Not found .env file")
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	// Gọi hàm setup router đã tách ra file riêng
	r := routers.SetupRouter()

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
		fmt.Println("Not found server port, default port: 8080")
	}

	if err := r.Run(":" + serverPort); err != nil {
		log.Fatal("Can't run server:", err)
	}

	fmt.Println("Server is running at http://localhost:" + serverPort)
}
