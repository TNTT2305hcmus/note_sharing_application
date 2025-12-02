package main

import (
	"fmt"
	"log"
	"note_sharing_application/server/router"
	"note_sharing_application/server/utils"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Server is booting...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Lỗi: Không tìm thấy file .env.")
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := utils.GenerateServerRSAKeys(); err != nil {
		log.Fatal("Fail to generate Server_RSA_Key:", err)
	}
	fmt.Println("PrivKeyRSA\n: ", utils.ServerPrivateKey)
	fmt.Println("PubKeyRSA\n: ", utils.ServerPublicKey)

	// Gọi hàm setup router đã tách ra file riêng
	r := router.SetupRouter()

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // Mặc định chạy port 8080 nếu quên cấu hình
		fmt.Println("Không tìm thấy SERVER_PORT, sử dụng mặc định: 8080")
	}
	// Chạy server tại cổng 8080
	fmt.Println("Server đang chạy tại http://localhost: " + serverPort)

	// Chạy Server bằng HTTP
	if err := r.Run(":" + serverPort); err != nil {
		log.Fatal("Không thể khởi động server:", err)
	}
}
