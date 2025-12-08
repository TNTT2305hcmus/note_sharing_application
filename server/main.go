package main

import (
	"fmt"
	"log"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/routers"
	"note_sharing_application/server/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func printKeyInfo(title, keyContent string) {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("%s:\n", title)

	lines := strings.Split(keyContent, "\n")
	if len(lines) > 5 {
		// Chỉ in 2 dòng đầu và 2 dòng cuối nếu key quá dài
		fmt.Println(lines[0])
		fmt.Println(lines[1])
		fmt.Println("... (Đã ẩn bớt nội dung giữa) ...")
		fmt.Println(lines[len(lines)-2])
	} else {
		fmt.Println(keyContent)
	}
	fmt.Println("--------------------------------------------------")
}

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

	configs.ConnectDB()
	handlers.UserCollection = configs.GetCollection("users")

	if handlers.UserCollection == nil {
		log.Fatal("Lỗi: UserCollection chưa được khởi tạo!")
	} else {
		fmt.Println("Đã khởi tạo UserCollection thành công")
	}

	// Sinh khóa RSA
	fmt.Println("\nĐang khởi tạo hệ thống mật mã RSA...")
	if err := utils.GenerateServerRSAKeys(); err != nil {
		log.Fatal("Lỗi: Không thể sinh khóa RSA Server:", err)
	}

	// In khóa RSA của server ra màn hình
	printKeyInfo("Server Private Key", fmt.Sprintf("%v", utils.ServerPrivateKey))
	pubKeyPEM, _ := utils.ExportPublicKeyAsPEM()
	printKeyInfo("Server Public Key (PEM)", pubKeyPEM)
	fmt.Printf("\n")

	// Gọi hàm setup router đã tách ra file riêng
	r := routers.SetupRouter()

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		// Mặc định chạy port 8080 nếu quên cấu hình
		serverPort = "8080"
		fmt.Println("Không tìm thấy SERVER_PORT, sử dụng mặc định: 8080")
	}

	configs.ConnectDB()
	handlers.UserCollection = configs.GetCollection("users")

	if err := r.Run(":" + serverPort); err != nil {
		log.Fatal("Lỗi: Không thể khởi động server:", err)
	}
	// Chạy server tại cổng 8080
	fmt.Println("Server đang chạy tại http://localhost:" + serverPort)
}
