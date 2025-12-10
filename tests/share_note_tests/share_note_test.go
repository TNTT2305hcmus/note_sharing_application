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

const TestDBName = "NoteApp_E2E_CleanTest"

var router *gin.Engine
var testDB *mongo.Database

// Struct giả lập response từ server để test độc lập
type TestNoteResponse struct {
	EncryptedContent string `json:"cipher_text"`
	KeyOption1       string `json:"encrypted_aes_key"`
	KeyOption2       string `json:"shared_encrypted_aes_key"`
	KeyOption3       string `json:"encrypted_aes_key_by_K"`
}

// --- TESTS ---

// Chạy trước và sau toàn bộ test
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("DB Connection Error:", err)
	}

	testDB = client.Database(TestDBName)
	configs.DB = testDB
	handlers.UserCollection = testDB.Collection("users")

	_ = utils.GenerateServerRSAKeys()
	router = routers.SetupRouter()

	exitVal := m.Run()

	_ = testDB.Drop(context.Background())
	os.Exit(exitVal)
}

// Test toàn bộ flow chia sẻ ghi chú giữa 2 user
func TestEndToEndSharingFlow(t *testing.T) {
	// 1. Setup Data
	alice, bob, password := "alice_clean", "bob_clean", "Pass123"
	noteContent := "Secret Content for Bob"

	// 2. Auth Flow
	alicePriv, _, _ := registerUser(t, alice, password)
	aliceToken, _ := loginUser(t, alice, password)

	bobPriv, _, bobPubHex := registerUser(t, bob, password)
	bobToken, _ := loginUser(t, bob, password)

	// 3. Alice Create Note
	aesKey, _ := crypto.GenerateAESKey()

	// Mã hóa nội dung
	encContent, _ := crypto.EncryptBytes([]byte(noteContent), aesKey)
	cipherTextBase64 := base64.StdEncoding.EncodeToString(encContent)

	// Mã hóa AES Key bằng Pass
	encAesKeyByPass, _ := crypto.EncryptByPassword(hex.EncodeToString(aesKey), password)

	noteID := createNoteRequest(t, aliceToken, cipherTextBase64, encAesKeyByPass)
	assert.NotEmpty(t, noteID)

	// 4. Alice Share to Bob
	sharedK, _ := crypto.ComputeSharedSecret(alicePriv, bobPubHex)
	sharedEncAesKey, _ := crypto.EncryptAESKeyWithSharedK(aesKey, sharedK)

	shareReq := models.CreateUrlRequest{
		SharedEncryptedAESKey: sharedEncAesKey,
		ExpiresIn:             "1h",
		MaxAccess:             5,
		Receiver:              bob,
	}
	_ = shareNoteRequest(t, aliceToken, noteID, shareReq)

	// 5. Bob Receive
	// Lấy danh sách link được share
	receivedUrls := getReceivedUrlsRequest(t, bobToken)
	var targetUrlID string
	for _, u := range receivedUrls {
		if u.NoteID == noteID {
			targetUrlID = u.ID.Hex() // Convert ObjectID -> String
			break
		}
	}
	assert.NotEmpty(t, targetUrlID, "Bob không thấy file được share")

	// Lấy chi tiết note
	noteData := viewNoteRequest(t, bobToken, targetUrlID)

	// 6. Bob Decrypt
	// Tính lại Shared Secret (Bob Priv + Alice Pub)
	alicePubHex := getPubKeyRequest(t, alice)
	bobSharedK, _ := crypto.ComputeSharedSecret(bobPriv, alicePubHex)

	// Lấy Key từ JSON (Kiểm tra 3 trường hợp tên field)
	targetKey := noteData.KeyOption3 // Ưu tiên 'encrypted_aes_key_by_K'
	if targetKey == "" {
		targetKey = noteData.KeyOption2
	}
	if targetKey == "" {
		targetKey = noteData.KeyOption1
	}

	assert.NotEmpty(t, targetKey, "Không tìm thấy AES Key trong JSON trả về")

	// Giải mã AES Key
	decryptedAESKey, err := crypto.DecryptAESKeyWithSharedK(targetKey, bobSharedK)
	assert.NoError(t, err, "Lỗi giải mã Key (Check lại Shared Secret hoặc Key Hex)")
	assert.Equal(t, aesKey, decryptedAESKey, "AES Key sau giải mã không khớp")

	// Giải mã nội dung
	cipherBytes, _ := base64.StdEncoding.DecodeString(noteData.EncryptedContent)
	decryptedContent, err := crypto.DecryptBytes(cipherBytes, decryptedAESKey)

	assert.NoError(t, err)
	assert.Equal(t, noteContent, string(decryptedContent))

	fmt.Printf("\nSUCCESS: '%s' == '%s'\n", noteContent, string(decryptedContent))
}

// --- HELPERS ---

// Mã hóa password bằng RSA Public Key của Server
func encryptPass(raw string) string {
	rng := rand.Reader
	enc, _ := rsa.EncryptOAEP(sha256.New(), rng, utils.ServerPublicKey, []byte(raw), []byte(""))
	return base64.StdEncoding.EncodeToString(enc)
}

// Đăng ký user, trả về Private Key, Public Key và Public Key Hex
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
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	return priv, pub, pubHex
}

// Đăng nhập user, trả về Token và Encrypted Private Key
func loginUser(t *testing.T, user, pass string) (string, string) {
	body, _ := json.Marshal(map[string]string{
		"username": user, "password": encryptPass(pass),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["token"].(string), res["encrypted_privKey"].(string)
}

// Tạo Note Request
func createNoteRequest(t *testing.T, token, cipher, encKey string) string {
	body, _ := json.Marshal(models.CreateNoteRequest{
		CipherText: cipher, EncryptedAesKey: encKey,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["note_id"]
}

// Chia sẻ Note Request
func shareNoteRequest(t *testing.T, token, noteID string, body models.CreateUrlRequest) string {
	jsonVal, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/notes/%s/url", noteID), bytes.NewBuffer(jsonVal))
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	return ""
}

// Lấy danh sách URL được share đến user
func getReceivedUrlsRequest(t *testing.T, token string) []models.Url {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notes/received", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	var urls []models.Url
	_ = json.Unmarshal(w.Body.Bytes(), &urls)
	return urls
}

// Lấy Public Key của user
func getPubKeyRequest(t *testing.T, user string) string {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/users/"+user+"/pubkey", nil)
	router.ServeHTTP(w, req)
	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["public_key"]
}

// Xem chi tiết Note qua URL
func viewNoteRequest(t *testing.T, token, urlID string) TestNoteResponse {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/note/"+urlID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var note TestNoteResponse
	_ = json.Unmarshal(w.Body.Bytes(), &note)
	return note
}
