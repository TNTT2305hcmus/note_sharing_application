package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"note_sharing_application/client/crypto"
	"note_sharing_application/server/configs"
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/models"
	"note_sharing_application/server/routers"
	"note_sharing_application/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- CẤU HÌNH ---
const E2E_TestDBName = "E2E_Test"

var E2E_router *gin.Engine
var E2E_testDB *mongo.Database

// Struct hứng response khi xem chi tiết Note
type TestNoteResponseDTO struct {
	EncryptedContent      string `json:"cipher_text"`
	SharedEncryptedAESKey string `json:"encrypted_aes_key_by_K"` // JSON tag khớp với ViewNoteHandler
	Sender                string `json:"sender"`
}

// Struct hứng response danh sách URL
type TestUrlResponseDTO struct {
	ID       string `json:"url_id"` // Dùng string để nhận cả ObjectID hex và chuỗi
	NoteID   string `json:"note_id"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

// --- MAIN ENTRY POINT ---

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Thiết lập môi trường
	cleanup := setupTestEnvironment()

	// Chạy test
	exitVal := m.Run()

	// Dọn dẹp
	cleanup()
	os.Exit(exitVal)
}

// --- HÀM KHỞI TẠO MÔI TRƯỜNG ---
// Đảm bảo Router, DB và Keys luôn sẵn sàng trước khi test chạy
func setupTestEnvironment() func() {
	if E2E_router != nil && E2E_testDB != nil {
		return func() {}
	}

	fmt.Println("[INFO] Dang thiet lap moi truong Test...")

	// 1. Kết nối DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		cancel()
		log.Fatal("[ERROR] Loi ket noi DB:", err)
	}

	E2E_testDB = client.Database(E2E_TestDBName)
	configs.DB = E2E_testDB
	handlers.UserCollection = E2E_testDB.Collection("users")

	// 2. Khởi tạo RSA Keys (In-memory) cho Server Utils
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		cancel()
		log.Fatal("[ERROR] Khong the tao RSA Key:", err)
	}
	utils.ServerPrivateKey = priv
	utils.ServerPublicKey = &priv.PublicKey

	// 3. Khởi tạo Router
	E2E_router = routers.SetupRouter()
	if E2E_router == nil {
		cancel()
		log.Fatal("[ERROR] Router khoi tao that bai (nil)")
	}

	// Trả về hàm cleanup
	return func() {
		fmt.Println("[INFO] Dang don dep du lieu Test...")
		_ = E2E_testDB.Drop(context.Background())
		_ = client.Disconnect(context.Background())
		cancel()
	}
}

// --- TEST CASES ---

func TestEndToEndSharingFlow(t *testing.T) {
	// Gọi hàm setup để đảm bảo môi trường đã sẵn sàng (phòng trường hợp chạy test đơn lẻ)
	_ = setupTestEnvironment()

	if E2E_router == nil {
		t.Fatal("[FATAL] E2E_router is nil")
	}

	// 1. Chuẩn bị dữ liệu
	alice, bob, password := "alice_clean", "bob_clean", "Pass123"
	noteContent := "Secret Content for Bob"

	// 2. Quy trình xác thực (Đăng ký/Đăng nhập)
	alicePriv, _, _ := registerUser(t, alice, password)
	aliceToken, _ := loginUser(t, alice, password)

	bobPriv, _, bobPubHex := registerUser(t, bob, password)
	bobToken, _ := loginUser(t, bob, password)

	// 3. Alice tạo ghi chú
	aesKey, _ := crypto.GenerateAESKey()

	// Mã hóa nội dung và Key (bằng password của Alice)
	encContent, _ := crypto.EncryptBytes([]byte(noteContent), aesKey)
	cipherTextBase64 := base64.StdEncoding.EncodeToString(encContent)
	encAesKeyByPass, _ := crypto.EncryptByPassword(hex.EncodeToString(aesKey), password)

	noteID := createNoteRequest(t, aliceToken, cipherTextBase64, encAesKeyByPass)
	assert.NotEmpty(t, noteID, "NoteID khong duoc rong")

	// 4. Alice chia sẻ cho Bob
	// Tính Shared Secret (Alice Priv + Bob Pub)
	sharedK, _ := crypto.ComputeSharedSecret(alicePriv, bobPubHex)
	// Mã hóa AES Key bằng Shared Secret
	sharedEncAesKey, _ := crypto.EncryptAESKeyWithSharedK(aesKey, sharedK)

	shareReq := models.CreateUrlRequest{
		SharedEncryptedAESKey: sharedEncAesKey,
		ExpiresIn:             "1h",
		MaxAccess:             5,
		Receiver:              bob,
		Sender:                alice,
	}
	_ = shareNoteRequest(t, aliceToken, noteID, shareReq)

	// 5. Bob nhận liên kết chia sẻ
	receivedUrls := getReceivedUrlsRequest(t, bobToken)
	var targetUrlID string

	// Logic tách ID từ URL (Server trả về localhost:8080/note/ID -> Lấy ID cuối)
	for _, u := range receivedUrls {
		if u.NoteID == noteID {
			parts := strings.Split(u.ID, "/")
			if len(parts) > 0 {
				targetUrlID = parts[len(parts)-1]
			} else {
				targetUrlID = u.ID
			}
			break
		}
	}
	assert.NotEmpty(t, targetUrlID, "Bob khong tim thay link duoc share")

	// 6. Bob xem chi tiết ghi chú
	noteData := viewNoteRequest(t, bobToken, targetUrlID)

	// 7. Bob giải mã dữ liệu
	// Tính Shared Secret (Bob Priv + Alice Pub)
	alicePubHex := getPubKeyRequest(t, alice)
	bobSharedK, _ := crypto.ComputeSharedSecret(bobPriv, alicePubHex)

	// Lấy Key đã mã hóa từ response DTO
	targetKey := noteData.SharedEncryptedAESKey
	assert.NotEmpty(t, targetKey, "Response JSON thieu truong 'encrypted_aes_key_by_K'")

	// Giải mã AES Key
	decryptedAESKey, err := crypto.DecryptAESKeyWithSharedK(targetKey, bobSharedK)
	assert.NoError(t, err, "Loi giai ma AES Key")
	assert.Equal(t, aesKey, decryptedAESKey, "AES Key sau giai ma khong khop")

	// Giải mã nội dung
	cipherBytes, _ := base64.StdEncoding.DecodeString(noteData.EncryptedContent)
	decryptedContent, err := crypto.DecryptBytes(cipherBytes, decryptedAESKey)

	assert.NoError(t, err, "Loi giai ma noi dung ghi chu")
	assert.Equal(t, noteContent, string(decryptedContent), "Noi dung giai ma sai lech")

	fmt.Printf("\n[SUCCESS] Original: '%s' == Decrypted: '%s'\n", noteContent, string(decryptedContent))
}

// --- HELPER FUNCTIONS ---

func encryptPass(raw string) string {
	rng := rand.Reader
	// utils.ServerPublicKey được đảm bảo tồn tại bởi setupTestEnvironment
	enc, err := rsa.EncryptOAEP(sha256.New(), rng, utils.ServerPublicKey, []byte(raw), []byte(""))
	if err != nil {
		log.Fatal("[ERROR] Loi ma hoa password:", err)
	}
	return base64.StdEncoding.EncodeToString(enc)
}

func registerUser(t *testing.T, user, pass string) (*big.Int, *big.Int, string) {
	priv, pub, _ := crypto.GenerateKeyPair()
	privHex, pubHex := priv.Text(16), pub.Text(16)
	encPriv, _ := crypto.EncryptByPassword(privHex, pass)

	body, _ := json.Marshal(map[string]string{
		"username": user, "password": encryptPass(pass),
		"public_key": pubHex, "encrypted_privKey": encPriv,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	E2E_router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	return priv, pub, pubHex
}

func loginUser(t *testing.T, user, pass string) (string, string) {
	body, _ := json.Marshal(map[string]string{
		"username": user, "password": encryptPass(pass),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	E2E_router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["token"].(string), res["encrypted_privKey"].(string)
}

func createNoteRequest(t *testing.T, token, cipher, encKey string) string {
	// Sử dụng model server cho request body để đảm bảo đúng định dạng
	reqBody := models.CreateNoteRequest{
		CipherText:      cipher,
		EncryptedAesKey: encKey,
		Sender:          "alice_clean",
	}

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	E2E_router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["note_id"]
}

func shareNoteRequest(t *testing.T, token, noteID string, body models.CreateUrlRequest) string {
	jsonVal, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/notes/%s/url", noteID), bytes.NewBuffer(jsonVal))
	req.Header.Set("Authorization", "Bearer "+token)
	E2E_router.ServeHTTP(w, req)
	assert.True(t, w.Code == 200 || w.Code == 201)
	return ""
}

func getReceivedUrlsRequest(t *testing.T, token string) []TestUrlResponseDTO {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/received", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	E2E_router.ServeHTTP(w, req)

	var urls []TestUrlResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &urls)
	assert.NoError(t, err)
	return urls
}

func getPubKeyRequest(t *testing.T, user string) string {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/users/"+user+"/pubkey", nil)
	E2E_router.ServeHTTP(w, req)
	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["public_key"]
}

func viewNoteRequest(t *testing.T, token, urlID string) TestNoteResponseDTO {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/note/"+urlID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	E2E_router.ServeHTTP(w, req)

	if w.Code != 200 {
		fmt.Printf("[DEBUG] ViewNote Error: Code=%d Body=%s\n", w.Code, w.Body.String())
	}
	assert.Equal(t, 200, w.Code)

	var note TestNoteResponseDTO
	_ = json.Unmarshal(w.Body.Bytes(), &note)
	return note
}
