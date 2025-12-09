package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"note_sharing_application/client/crypto"
	"note_sharing_application/client/models"
	"note_sharing_application/client/services"
)

// Struct để lưu thông tin phiên làm việc
type Session struct {
	Username            string `json:"username"`
	Token               string `json:"token"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

func printHelp() {
	fmt.Println("\n------------------------ ỨNG DỤNG CHIA SẺ GHI CHÚ BẢO MẬT (CLI) -------------------------------")
	fmt.Println("1. Đăng ký:		go run main.go register -u <user> -p <pass>")
	fmt.Println("2. Đăng nhập:  	go run main.go login -u <user> -p <pass>")
	fmt.Println("3. Liệt kê file cá nhân:            go run main.go listOwnedFile -u <current username>")
	fmt.Println("4. Liệt kê file được chia sẻ:       go run main.go listSharedFile -u <current username>")
	fmt.Println("5. Lưu file mã hóa lên server:      go run main.go save -f <path> -u <current username>")
	fmt.Println("6. Gửi file (Chia sẻ):              go run main.go send -note <id> -t <receiver> [-exp 1h] [-max 1] -u <current username>")
	fmt.Println("7. Xóa file gốc:                    go run main.go deleteFile -id <id> -u <current username>")
	fmt.Println("8. Hủy chia sẻ:                     go run main.go cancelSharingURL -id <id> -u <current username>")
	fmt.Println("9. Đọc ghi chú được chia sẻ:        go run main.go readSharedNote -id <url_id> -sender <sender_name> -u <current username> -o <output_file>")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}
	switch os.Args[1] {

	case "register":
		registerCmd := flag.NewFlagSet("register", flag.ExitOnError)
		regUser := registerCmd.String("u", "", "Username")
		regPass := registerCmd.String("p", "", "Password")
		registerCmd.Parse(os.Args[2:])
		handleRegister(*regUser, *regPass)

	case "login":
		loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
		loginUser := loginCmd.String("u", "", "Username")
		loginPass := loginCmd.String("p", "", "Password")
		loginCmd.Parse(os.Args[2:])
		handleLogin(*loginUser, *loginPass)

	case "listOwnedFile":
		cmd := flag.NewFlagSet("listOwnedFile", flag.ExitOnError)
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleListOwnedFile(*user)

	case "listSharedFile":
		cmd := flag.NewFlagSet("listSharedFile", flag.ExitOnError)
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleListSharedFile(*user)

	case "save":
		cmd := flag.NewFlagSet("save", flag.ExitOnError)
		filePath := cmd.String("f", "", "File path")
		// Thêm cờ -u để biết ai đang save
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleSaveFile(*filePath, *user)

	case "send":
		cmd := flag.NewFlagSet("send", flag.ExitOnError)
		noteID := cmd.String("note", "", "Note ID")
		receiver := cmd.String("t", "", "Receiver")
		expiresIn := cmd.String("exp", "24h", "Expire")
		maxAccess := cmd.Int("max", 1, "Max Access")
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleSendFile(*noteID, *receiver, *expiresIn, *maxAccess, *user)

	case "deleteFile":
		// Cú pháp: deleteFile -id <note_id>
		cmd := flag.NewFlagSet("deleteFile", flag.ExitOnError)
		noteID := cmd.String("id", "", "ID của ghi chú cần xóa")
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleDeleteFile(*noteID, *user)

	case "cancelSharingURL":
		// Cú pháp: cancelSharingURL -id <note_id>
		cmd := flag.NewFlagSet("cancelSharingURL", flag.ExitOnError)
		noteID := cmd.String("id", "", "ID của ghi chú muốn hủy chia sẻ")
		user := cmd.String("u", "", "Current username")
		cmd.Parse(os.Args[2:])
		handleCancelSharing(*noteID, *user)

	case "readSharedNote":
		// Cú pháp: readSharedNote -id <url_id> -sender <sender> -o <path> -u <me>
		cmd := flag.NewFlagSet("readSharedNote", flag.ExitOnError)

		urlID := cmd.String("id", "", "ID của URL chia sẻ (Lấy từ listSharedFile)")
		sender := cmd.String("sender", "", "Username người gửi (Lấy từ listSharedFile)")
		outFile := cmd.String("o", "", "Đường dẫn file để lưu kết quả giải mã")
		user := cmd.String("u", "", "Username của bạn")

		cmd.Parse(os.Args[2:])
		handleReadSharedNote(*urlID, *sender, *outFile, *user)

	default:
		printHelp()
	}
}

// --- CÁC HÀM XỬ LÝ LOGIC ---

func handleRegister(user, pass string) {
	if user == "" || pass == "" {
		fmt.Println("Lỗi: Thiếu thông tin.")
		fmt.Println("VD: go run main.go register -u alice -p 123")
		return
	}

	fmt.Println("Đang sinh cặp khóa Diffie-Hellman...")
	privKey, pubKey, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Println("Lỗi: Không thể sinh khóa Diffie-Hellman:", err)
		return
	}

	// Chuyển sang Hex để gửi và mã hóa
	privKeyHex := privKey.Text(16)
	pubKeyHex := pubKey.Text(16)
	fmt.Printf("Public Key sinh ra: %s...\n", pubKeyHex[:10])

	fmt.Println("Đang mã hóa Private Key bằng Password...")
	encryptedPrivKey, err := crypto.EncryptByPassword(privKeyHex, pass)
	if err != nil {
		fmt.Println("Lỗi: Không thể mã hóa Private Key:", err)
		return
	}

	fmt.Println("Đang gọi API Đăng ký...")
	err = services.Register(user, pass, pubKeyHex, encryptedPrivKey)
	if err != nil {
		fmt.Println("Lỗi: Đăng ký thất bại:", err)
		return
	}

	fmt.Println("Đăng ký thành công")
}

func handleLogin(user, pass string) {
	if user == "" || pass == "" {
		fmt.Println("Lỗi: Thiếu thông tin.")
		fmt.Println("VD: go run main.go register -u alice -p 123")
		return
	}

	fmt.Println("Đang gọi API Đăng nhập...")
	token, encryptedPrivKey, err := services.Login(user, pass)
	if err != nil {
		fmt.Println("Lỗi: Đăng nhập thất bại:", err)
		return
	}
	fmt.Println("Đăng nhập thành công.")

	// Lưu Token và EncryptedPrivateKey vào file
	saveSession(Session{
		Username:            user,
		Token:               token,
		EncryptedPrivateKey: encryptedPrivKey,
	})
	fmt.Println("Đã lưu phiên làm việc.")
}

func handleListOwnedFile(username string) {
	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	notes, err := services.GetOwnedNotes(session.Token)
	if err != nil {
		fmt.Printf("Lỗi: Không thể lấy danh sách file: %v\n", err)
		return
	}

	fmt.Println("\n--- DANH SÁCH FILE CỦA BẠN ---")
	if len(notes) == 0 {
		fmt.Println("(Trống)")
		return
	}
	for _, n := range notes {
		fmt.Printf("- Note ID: %s\n", n.ID)
	}
}

func handleListSharedFile(username string) {
	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	urls, err := services.GetReceivedURLs(session.Token)
	if err != nil {
		fmt.Printf("Lỗi: Không thể lấy danh sách chia sẻ: %v\n", err)
		return
	}

	fmt.Println("\n--- DANH SÁCH ĐƯỢC CHIA SẺ VỚI BẠN ---")
	if len(urls) == 0 {
		fmt.Println("(Trống)")
		return
	}
	for _, u := range urls {
		fmt.Printf("- URL ID: %s | Từ: %s | Note ID: %s | Hết hạn: %v\n",
			u.ID, u.SenderID, u.NoteID, u.ExpiresAt)
	}
}

func handleSaveFile(filePath, username string) {
	if filePath == "" {
		fmt.Println("Vui lòng nhập đường dẫn file: -f <path>")
		return
	}
	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	// Cần mật khẩu để mã hóa AES Key
	password := promptPassword("Nhập mật khẩu để mã hóa khóa file: ")

	fmt.Println("Đang xử lý mã hóa file...")
	// Sử dụng crypto package để mã hóa file và bọc khóa AES bằng password
	cipherTextBase64, encryptedAESKey, err := crypto.PrepareFileForUpload(filePath, password)
	if err != nil {
		fmt.Printf("Lỗi mã hóa local: %v\n", err)
		return
	}

	// Upload lên server
	noteID, err := services.CreateNote(session.Token, cipherTextBase64, encryptedAESKey)
	if err != nil {
		fmt.Printf("Lỗi upload lên server: %v\n", err)
		return
	}

	fmt.Printf("Lưu thành công! Note ID: %s\n", noteID)
}

// Logic:
// B1. Lấy EncryptedAESKeyByPass của Note  -> Giải mã bằng Pass.
// B2. Lấy PubKey của Receiver -> Tính Shared Secret K (Diffie-Hellman).
// B3. Mã hóa AES Key bằng K -> Gửi lên Server tạo URL.
func handleSendFile(noteID, receiver, expiresIn string, maxAccess int, username string) {
	if noteID == "" || receiver == "" {
		fmt.Println("Thiếu thông tin. Cần: -note <id> -t <receiver>")
		return
	}

	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}
	// Nhập mật khẩu để giải mã EncryptedPrivKey và EncryptedAESKey
	password := promptPassword("Nhập mật khẩu xác thực: ")

	// Tìm Note để lấy EncryptedAesKey được mã bằng password
	myNotes, err := services.GetOwnedNotes(session.Token)
	if err != nil {
		fmt.Println("Lỗi lấy danh sách note:", err)
		return
	}
	var targetNote *models.Note
	for _, n := range myNotes {
		if n.ID == noteID {
			targetNote = &n
			break
		}
	}
	if targetNote == nil {
		fmt.Println("Không tìm thấy Note ID này trong danh sách sở hữu của bạn.")
		return
	}

	// Giải mã EncryptedAESKey bằng password
	fmt.Println("Đang giải mã khóa AES gốc...")
	aesKeyRawHex, err := crypto.DecryptByPassword(targetNote.EncryptedAesKey, password)
	if err != nil {
		fmt.Println("Sai mật khẩu hoặc dữ liệu lỗi:", err)
		return
	}
	aesKeyBytes, _ := hex.DecodeString(aesKeyRawHex)

	// Diffie-Hellman
	// Giải mã EncryptedPrivKey bằng password
	fmt.Println("Đang giải mã Private Key DH của bạn...")
	myPrivKeyHex, err := crypto.DecryptByPassword(session.EncryptedPrivateKey, password)
	if err != nil {
		fmt.Println("Lỗi giải mã Private Key:", err)
		return
	}
	myPrivKeyBig := new(big.Int)
	myPrivKeyBig.SetString(myPrivKeyHex, 16)

	// Lấy Pubkey của Receiver từ Server
	fmt.Printf("Đang lấy Public Key của %s...\n", receiver)
	receiverPubKeyHex, err := services.GetUserPublicKey(receiver)
	if err != nil {
		fmt.Println("Lỗi lấy key người nhận (có thể user không tồn tại):", err)
		return
	}

	// Tính khóa chung K
	sharedK, err := crypto.ComputeSharedSecret(myPrivKeyBig, receiverPubKeyHex)
	if err != nil {
		fmt.Println("Lỗi tính khóa chung:", err)
		return
	}

	// Mã hóa AES Key bằng Shared K
	sharedEncryptedAESKey, err := crypto.EncryptAESKeyWithSharedK(aesKeyBytes, sharedK)
	if err != nil {
		fmt.Println("Lỗi mã hóa khóa chia sẻ:", err)
		return
	}

	// Gọi API tạo Share URL
	fmt.Println("Đang gửi yêu cầu chia sẻ lên server...")
	err = services.CreateNoteUrl(noteID, session.Token, sharedEncryptedAESKey, expiresIn, receiver, maxAccess, username)
	if err != nil {
		fmt.Println("Chia sẻ thất bại:", err)
		return
	}

	fmt.Println("Chia sẻ thành công! Người nhận có thể thấy trong danh sách của họ.")
}

func handleDeleteFile(noteID, username string) {
	if noteID == "" {
		fmt.Println("Thiếu Note ID: -id <id>")
		return
	}
	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	err = services.DeleteNote(session.Token, noteID)
	if err != nil {
		fmt.Println("Xóa thất bại:", err)
		return
	}
	fmt.Println("Đã xóa ghi chú vĩnh viễn.")
}

func handleCancelSharing(noteID, username string) {
	if noteID == "" {
		fmt.Println("Thiếu Note ID: -id <id>")
		return
	}
	if username == "" {
		fmt.Println("Vui lòng chỉ định user: -u <username>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	// Lưu ý: Hàm services.DeleteSharedNote cần ID của Note để xóa tất cả share liên quan
	err = services.DeleteSharedNote(session.Token, noteID)
	if err != nil {
		fmt.Println("Hủy chia sẻ thất bại:", err)
		return
	}
	fmt.Println("Đã hủy chia sẻ ghi chú này.")
}

// Logic:
// B1. Tải CipherText và EncryptedKey (bọc bởi K) từ Server.
// B2. Lấy PubKey của Sender -> Tính Shared Secret K.
// B3. Dùng K giải mã lấy AES Key gốc.
// B4. Dùng AES Key giải mã CipherText -> Ghi ra file.
func handleReadSharedNote(urlID, sender, outFile, username string) {
	// Kiểm tra đầu vào
	if urlID == "" || sender == "" || outFile == "" || username == "" {
		fmt.Println("Thiếu thông tin. Cần: -id <url_id> -sender <name> -o <path> -u <me>")
		return
	}

	session, err := loadSession(username)
	if err != nil {
		fmt.Println("Lỗi session:", err)
		return
	}

	// Nhập mật khẩu để giải mã EncryptedPrivKey
	password := promptPassword("Nhập mật khẩu của BẠN để giải mã: ")

	// Gọi API lấy ciphertext
	fmt.Println("Đang tải dữ liệu từ server...")
	noteData, err := services.ReadNoteWithURL(urlID, session.Token)
	if err != nil {
		fmt.Printf("Lỗi tải dữ liệu: %v\n", err)
		return
	}

	// Diffie-Hellman
	fmt.Println("Đang tính toán khóa chung (Shared Secret)...")

	// Giải mã Private Key
	myPrivKeyHex, err := crypto.DecryptByPassword(session.EncryptedPrivateKey, password)
	if err != nil {
		fmt.Println("Sai mật khẩu hoặc lỗi Private Key:", err)
		return
	}
	myPrivKeyBig := new(big.Int)
	myPrivKeyBig.SetString(myPrivKeyHex, 16)

	// Lấy Public Key của Sender
	senderPubKeyHex, err := services.GetUserPublicKey(sender)
	if err != nil {
		fmt.Printf("Lỗi lấy Public Key của %s: %v\n", sender, err)
		return
	}

	// Tính K
	sharedK, err := crypto.ComputeSharedSecret(myPrivKeyBig, senderPubKeyHex)
	if err != nil {
		fmt.Println("Lỗi tính toán Diffie-Hellman:", err)
		return
	}

	// Giải mã AES Key bằng K
	fmt.Println("Đang giải mã khóa AES...")
	fmt.Println(sender)
	aesKeyBytes, err := crypto.DecryptAESKeyWithSharedK(noteData.EncryptedKey, sharedK)
	if err != nil {
		fmt.Println("Giải mã khóa thất bại (Có thể sai Sender hoặc Token bị lỗi):", err)
		return
	}

	// Giải mã nội dung file bằng AES Key vừa tìm được
	fmt.Println("Đang giải mã nội dung ghi chú...")

	// Chuyển lại AES Key sang Hex string để tái sử dụng hàm RestoreFileFromNote cũ
	aesKeyHex := hex.EncodeToString(aesKeyBytes)

	err = crypto.RestoreFileFromNote(noteData.EncryptedContent, aesKeyHex, outFile)
	if err != nil {
		fmt.Println("Lỗi giải mã file:", err)
		return
	}

	fmt.Printf("Đã giải mã thành công!\nNội dung được lưu tại: %s\n", outFile)
}

// --- HÀM PHỤ TRỢ (Session) ---
// Hàm sinh tên file
func getSessionFilename(username string) string {
	// Nếu không truyền user, mặc định là session.json (fallback)
	if username == "" {
		return "session.json"
	}
	return fmt.Sprintf("session_%s.json", username)
}

func saveSession(s Session) {
	filename := getSessionFilename(s.Username)
	data, _ := json.Marshal(s)
	os.WriteFile(filename, data, 0644)
	fmt.Printf("💾 Đã lưu phiên làm việc của '%s' vào file: %s\n", s.Username, filename)
}

func loadSession(username string) (Session, error) {
	filename := getSessionFilename(username)
	data, err := os.ReadFile(filename)
	if err != nil {
		return Session{}, fmt.Errorf("không tìm thấy session của user '%s'. Bạn đã đăng nhập chưa?", username)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return Session{}, fmt.Errorf("file session lỗi")
	}
	return s, nil
}
func promptPassword(label string) string {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
