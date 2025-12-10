package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"note_sharing_application/server/configs"
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/routers"
	"note_sharing_application/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Đặt tên db test xác thực
const TestDBName = "NoteAppDB_Test_Auth"

var router *gin.Engine

func TestMain(m *testing.M) {
	// Set Gin mode Test
	gin.SetMode(gin.TestMode)

	// Kết nối MongoDB Test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Không thể kết nối MongoDB Test:", err)
	}

	configs.DB = client.Database(TestDBName)

	// Gán đè biến toàn cục trong handlers để dùng DB Test
	handlers.UserCollection = configs.DB.Collection("users")

	// Sinh khóa RSA cho Server
	if err := utils.GenerateServerRSAKeys(); err != nil {
		log.Fatal("Lỗi sinh RSA Key Server:", err)
	}

	// Setup Router
	router = routers.SetupRouter()

	// Chạy test
	exitVal := m.Run()

	// Xóa DB sau test
	fmt.Println("\n Đang dọn dẹp Database Test...")
	configs.DB.Drop(context.Background())

	os.Exit(exitVal)
}

// Mã hóa password bằng server pubkey RSA
func encryptPasswordForTest(rawPassword string) string {
	pubKey := utils.ServerPublicKey

	// Mã hóa OAEP
	rng := rand.Reader
	label := []byte("")
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rng, pubKey, []byte(rawPassword), label)
	if err != nil {
		log.Fatal("Lỗi mã hóa password trong test:", err)
	}

	// Trả về Base64
	return base64.StdEncoding.EncodeToString(encryptedBytes)
}

func TestAuthFlow(t *testing.T) {
	// Dữ liệu mẫu
	username := "user_auth"
	password := "pass"

	// Dữ liệu giả cho DH Key
	mockPubKey := "MockPublicKeyHexString"
	mockEncPrivKey := "MockEncryptedPrivateKeyHex"

	// Test Register
	t.Run("1. Register Success", func(t *testing.T) {
		encryptedPass := encryptPasswordForTest(password)
		reqBody := map[string]string{
			"username":          username,
			"password":          encryptedPass,
			"public_key":        mockPubKey,
			"encrypted_privKey": mockEncPrivKey,
		}
		jsonValue, _ := json.Marshal(reqBody)

		// Tạo Request giả
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		// Ghi lại Response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Kiểm tra kết quả
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully Registered")
	})

	// Lỗi đăng ký nếu trùng username
	t.Run("2. Register Duplicate Username (Fail)", func(t *testing.T) {
		// Gửi lại request y hệt user trên
		encryptedPass := encryptPasswordForTest(password)
		reqBody := map[string]string{
			"username":          username,
			"password":          encryptedPass,
			"public_key":        mockPubKey,
			"encrypted_privKey": mockEncPrivKey,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Lỗi 409 (StatusConflict)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	// Test Login
	t.Run("3. Login Success", func(t *testing.T) {
		encryptedPass := encryptPasswordForTest(password)
		reqBody := map[string]string{
			"username": username,
			"password": encryptedPass,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response để check token
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response["token"], "Token không được rỗng")
		assert.Equal(t, mockEncPrivKey, response["encrypted_privKey"], "Phải trả về Encrypted Private Key")
	})

	// Đăng nhập lỗi nếu sai password
	t.Run("4. Login Wrong Password (Fail)", func(t *testing.T) {
		encryptedPass := encryptPasswordForTest("WrongPass")
		reqBody := map[string]string{
			"username": username,
			"password": encryptedPass,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Đăng nhập lỗi nếu chưa đăng ký
	t.Run("5. Login User Not Found (Fail)", func(t *testing.T) {
		encryptedPass := encryptPasswordForTest(password)
		reqBody := map[string]string{
			"username": "ghost_user",
			"password": encryptedPass,
		}
		jsonValue, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
