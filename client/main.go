package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"

	"note_sharing_application/client/crypto"
	"note_sharing_application/client/services"
)

// Struct để lưu thông tin phiên làm việc
type Session struct {
	Username            string `json:"username"`
	Token               string `json:"token"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}
	// Các lệnh con
	// Lệnh: register -u <user> -p <pass>
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)
	regUser := registerCmd.String("u", "", "Username")
	regPass := registerCmd.String("p", "", "Password")

	// Lệnh: login -u <user> -p <pass>
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	loginUser := loginCmd.String("u", "", "Username")
	loginPass := loginCmd.String("p", "", "Password")

	// Lệnh: get-key -t <target_username> (Lấy Public Key của người khác)
	getKeyCmd := flag.NewFlagSet("get-key", flag.ExitOnError)
	targetUser := getKeyCmd.String("t", "", "Username người cần lấy Key")

	// Lệnh: go run main.go connect -u <current_user> -t <target_user> (Test tính khóa chung K)
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	connectUser := connectCmd.String("u", "", "Username của BẠN (người đang chạy lệnh)")
	connectTarget := connectCmd.String("t", "", "Username người muốn kết nối")

	// 3. Switch để xử lý từng lệnh
	switch os.Args[1] {

	case "register":
		registerCmd.Parse(os.Args[2:])
		handleRegister(*regUser, *regPass)

	case "login":
		loginCmd.Parse(os.Args[2:])
		handleLogin(*loginUser, *loginPass)

	case "get-key":
		getKeyCmd.Parse(os.Args[2:])
		handleConnectReceiver(*targetUser)

	case "connect":
		connectCmd.Parse(os.Args[2:])
		handleConnect(*connectUser, *connectTarget)

	case "help":
		printHelp()

	default:
		fmt.Println("Lệnh không tồn tại.")
		printHelp()
	}
}

// --- CÁC HÀM XỬ LÝ LOGIC ---

func handleRegister(user, pass string) {
	if user == "" || pass == "" {
		fmt.Println("Thiếu thông tin. VD: go run main.go register -u alice -p 123")
		return
	}

	fmt.Println("Đang sinh cặp khóa Diffie-Hellman...")
	privKey, pubKey, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Println("Lỗi sinh khóa:", err)
		return
	}

	// Chuyển sang Hex để gửi và mã hóa
	privKeyHex := privKey.Text(16)
	pubKeyHex := pubKey.Text(16)
	fmt.Printf("Public Key sinh ra: %s...\n", pubKeyHex[:10])

	fmt.Println("Đang mã hóa Private Key bằng Password...")
	encryptedPrivKey, err := crypto.EncryptPrivateKeyWithPassword(privKeyHex, pass)
	if err != nil {
		fmt.Println("Lỗi mã hóa Private Key:", err)
		return
	}

	fmt.Println("Đang gọi API Đăng ký...")
	err = services.Register(user, pass, pubKeyHex, encryptedPrivKey)
	if err != nil {
		fmt.Println("Đăng ký thất bại:", err)
		return
	}
}

func handleLogin(user, pass string) {
	if user == "" || pass == "" {
		fmt.Println("Thiếu thông tin.")
		return
	}

	fmt.Println("Đang gọi API Đăng nhập...")
	token, encryptedPrivKey, err := services.Login(user, pass)
	if err != nil {
		fmt.Println("Đăng nhập thất bại:", err)
		return
	}
	fmt.Println("Đăng nhập thành công.")

	// Lưu Token và EncryptedPrivateKey vào file
	saveSession(Session{
		Username:            user,
		Token:               token,
		EncryptedPrivateKey: encryptedPrivKey,
	})
	fmt.Println("Đã lưu phiên làm việc (Bao gồm khóa được bảo vệ).")
}

func handleConnectReceiver(targetUser string) {
	if targetUser == "" {
		fmt.Println("Thiếu username. VD: go run main.go get-key -t bob")
		return
	}

	pubKey, err := services.GetUserPublicKey(targetUser)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}
	fmt.Printf("Public Key của %s:\n%s\n", targetUser, pubKey)
}

func handleConnect(currentUser, targetUser string) {
	if currentUser == "" || targetUser == "" {
		fmt.Println("Thiếu thông tin. VD: go run main.go connect -u alice -t bob")
		return
	}

	// Load session của ĐÚNG người dùng này
	session, err := loadSession(currentUser)
	if err != nil {
		return
	}

	// --- BƯỚC BẢO MẬT: HỎI PASSWORD ĐỂ GIẢI MÃ ---
	fmt.Printf("Để dùng Private Key, vui lòng nhập mật khẩu: ")
	var password string
	fmt.Scanln(&password)

	fmt.Println("Đang giải mã Private Key trong bộ nhớ tạm...")
	// Giải mã từ chuỗi Encrypted lưu trong file session
	privKeyHex, err := crypto.DecryptPrivateKeyWithPassword(session.EncryptedPrivateKey, password)
	if err != nil {
		fmt.Println("Sai mật khẩu! Không thể giải mã khóa.", err)
		return
	}

	fmt.Printf("Đang lấy Public Key của '%s'...\n", targetUser)
	peerPubKeyHex, err := services.GetUserPublicKey(targetUser)
	if err != nil {
		fmt.Println("Lỗi lấy key đối phương:", err)
		return
	}

	// Khôi phục Private Key BigInt
	myPrivKey := new(big.Int)
	myPrivKey.SetString(privKeyHex, 16)

	// Tính Shared Secret
	sharedKey, err := crypto.ComputeSharedSecret(myPrivKey, peerPubKeyHex)
	if err != nil {
		fmt.Println("Lỗi tính toán:", err)
		return
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("TÍNH KHÓA K THÀNH CÔNG!\n")
	fmt.Printf("Shared Secret: %s\n", sharedKey.Text(16))
	fmt.Println("------------------------------------------------")

}

// --- HÀM PHỤ TRỢ (Session) ---

// Helper sinh tên file
func getSessionFilename(username string) string {
	return fmt.Sprintf("session_%s.json", username)
}

func saveSession(s Session) {
	filename := getSessionFilename(s.Username)
	data, _ := json.Marshal(s)
	os.WriteFile(filename, data, 0644)
	fmt.Printf("Đã lưu phiên làm việc vào file: %s\n", filename)
}

// Hàm loadSession bây giờ cần tham số username để biết load file nào
func loadSession(username string) (Session, error) {
	filename := getSessionFilename(username)
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Không tìm thấy phiên làm việc của '%s'. Hãy đăng nhập trước.\n", username)
		return Session{}, err
	}
	var s Session
	json.Unmarshal(data, &s)
	return s, nil
}

func printHelp() {
	fmt.Println("\n--- ỨNG DỤNG CHIA SẺ GHI CHÚ BẢO MẬT (CLI) ---")
	fmt.Println("1. Đăng ký:       go run main.go register -u <user> -p <pass>")
	fmt.Println("2. Đăng nhập:     go run main.go login -u <user> -p <pass>")
	fmt.Println("3. Lấy public key của người nhận:  go run main.go get-key -t <target_user>")
	fmt.Println("4. Tính khóa K:   go run main.go connect -u <current_user> -t <target_user>")
}
