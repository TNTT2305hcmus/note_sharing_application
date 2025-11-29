package main

import (
	"bufio"
	"fmt"
	"note_sharing_application/client/crypto"
	"note_sharing_application/client/services"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- CLIENT APP ---")

	// Nhập thông tin người dùng
	fmt.Print("Nhập Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Nhập Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Client tự sinh cặp khóa (Private/Public) dựa trên G và P
	fmt.Println("Đang sinh cặp khóa...")
	privKey, pubKey, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Println("Lỗi sinh khóa:", err)
		return
	}

	// Chuyển Public Key sang String để gửi đi (Hex)
	pubKeyStr := pubKey.String()
	fmt.Printf("Public Key: %s\n", pubKeyStr)

	// Gửi yêu cầu Đăng Ký
	fmt.Println("\n--- Gửi yêu cầu Đăng Ký ---")
	err = services.Register(username, password, pubKeyStr)
	if err != nil {
		fmt.Println("Lỗi:", err)
		// Nếu muốn test login ngay cả khi user đã tồn tại thì không return ở đây
	}

	// 4. Gửi yêu cầu Đăng Nhập
	fmt.Println("\n--- Gửi yêu cầu Đăng Nhập ---")
	token, err := services.Login(username, password)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	// In tạm token + private key ở đây để khỏi lỗi
	fmt.Println("\nToken đã lưu:", token)
	fmt.Println("\nPrivate key:", privKey)
}
