package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"note_sharing_application/server/router"
	"note_sharing_application/server/utils"
)

func main() {
	// --- SETUP GIAO DIỆN CONSOLE ---
	// Xóa màn hình cho sạch (tùy chọn)
	fmt.Print("\033[H\033[2J")
	printBanner()

	// --- LOAD MÔI TRƯỜNG ---
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Cảnh báo: Không tìm thấy file .env (Sẽ dùng biến môi trường hệ thống)")
	} else {
		fmt.Println("Đã load file .env thành công")
	}

	// Setup Gin Mode
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor() // Bắt buộc hiện màu để dễ nhìn lỗi
	}

	// --- SINH KHÓA RSA SERVER ---
	fmt.Println("\nĐang khởi tạo hệ thống mật mã RSA...")
	if err := utils.GenerateServerRSAKeys(); err != nil {
		log.Fatal("Lỗi nghiêm trọng: Không thể sinh khóa RSA Server:", err)
	}

	// In khóa ra màn hình (Cắt ngắn để đỡ rối mắt, nhưng vẫn đủ để check)
	printKeyInfo("Server Private Key", fmt.Sprintf("%v", utils.ServerPrivateKey))

	// Xuất Public Key dạng PEM để dễ nhìn
	pubKeyPEM, _ := utils.ExportPublicKeyAsPEM()
	printKeyInfo("Server Public Key (PEM)", pubKeyPEM)

	// --- 4. TÙY CHỈNH LOGGING (Để quan sát Client) ---
	// Cấu hình format log để khi Client gọi API, nó hiện rõ: [GIỜ] | TRẠNG THÁI | THỜI GIAN | IP | METHOD | PATH
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		fmt.Printf("Route: %-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// --- 5. KHỞI ĐỘNG ROUTER ---
	fmt.Println("\nĐang thiết lập Router và Database...")
	r := router.SetupRouter()

	// --- 6. CHẠY SERVER ---
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
		fmt.Println("SERVER_PORT chưa set, sử dụng mặc định: 8080")
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("SERVER ĐANG CHẠY TẠI: http://localhost: %s\n", serverPort)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Sẵn sàng nhận yêu cầu từ Client...\n")

	if err := r.Run(":" + serverPort); err != nil {
		log.Fatal("Không thể khởi động server:", err)
	}
}

// --- CÁC HÀM PHỤ TRỢ ĐỂ IN ĐẸP ---

func printBanner() {
	fmt.Println(`
   ______                          
  / ____/___  _________  ___  _____
 / / __/ __ \/ ___/ __ \/ _ \/ ___/
/ /_/ / /_/ / /  / /_/ /  __/ /    
\____/ .___/_/  / .___/\___/_/     
    /_/        /_/                 
   SECURE NOTE SHARING BACKEND v1.0
	`)
}

func printKeyInfo(title, keyContent string) {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🔑 %s:\n", title)

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
