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
	// Chuyển privKey, pubKey từ big.Int sang Hex
	privKeyHex := privKey.Text(16)
	pubKeyHex := pubKey.Text(16)

	// Mã hóa privKey bằng password
	encryptedPrivKey, err := crypto.EncryptPrivateKeyWithPassword(privKeyHex, password)

	// Gửi yêu cầu Đăng Ký
	fmt.Println("\n--- Gửi yêu cầu Đăng Ký ---")
	err = services.Register(username, password, pubKeyHex, encryptedPrivKey)

	// 4. Gửi yêu cầu Đăng Nhập
	fmt.Println("\n--- Gửi yêu cầu Đăng Nhập ---")
	token, encryptedPrivKey, err := services.Login(username, password)

	// In tạm token + private key ở đây để khỏi lỗi
	fmt.Println("\nToken đã lưu:", token)
	fmt.Println("\nPrivate key đã mã hóa:", encryptedPrivKey)
}
