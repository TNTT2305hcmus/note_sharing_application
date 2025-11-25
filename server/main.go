package main

import (
	"fmt"
	"note_sharing_application/server/router"
)

func main() {
	fmt.Println("Server is booting...")

	// Gọi hàm setup router đã tách ra file riêng
	r := router.SetupRouter()

	// Chạy server tại cổng 8080
	r.Run("localhost:8080")
}

