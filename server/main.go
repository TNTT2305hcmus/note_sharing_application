package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"note_sharing_application/server/router"
	"os"
)

func main() {
	fmt.Println("Server is booting...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Lỗi: Không tìm thấy file .env.")
	}

	// Gọi hàm setup router đã tách ra file riêng
	r := router.SetupRouter()

	serverPort := os.Getenv("SERVER_PORT")
	// Chạy server tại cổng 8080
	fmt.Println("Server đang chạy tại port " + serverPort)

	if err := r.Run(":" + serverPort); err != nil {
		log.Fatal("Không thể khởi động server:", err)
	}
}
