package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"note_sharing_application/client/crypto"
	"note_sharing_application/client/models"

	"github.com/stretchr/testify/assert"
)

func SetupMockUser(t *testing.T, username, password string) string {
	// Giả lập đăng ký & đăng nhập để lấy token
	mockPubKey := "MockPublicKeyHexString"
	mockEncPrivKey := "MockEncryptedPrivateKeyHex"
	//Đăng ký
	encryptedPass := encryptPasswordForTest(password)
	reqBody := map[string]string{
		"username":          username,
		"password":          encryptedPass,
		"public_key":        mockPubKey,
		"encrypted_privKey": mockEncPrivKey,
	}
	jsonValue1, _ := json.Marshal(reqBody)

	// Tạo Request giả
	req1, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue1))
	req1.Header.Set("Content-Type", "application/json")

	// Ghi lại Response
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	//Đăng nhập để nhận token về test
	logBody := map[string]string{
		"username": username,
		"password": encryptedPass,
	}
	logJson, _ := json.Marshal(logBody)

	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(logJson))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Lỗi parse JSON:", err)
	}

	// Lấy ra và ép kiểu
	token, _ := response["token"].(string)
	return token
}

// Tạo Note test
func SetupMockNote(t *testing.T, password, token string) string {
	cipherTextBase64, encryptedAESKey, _ := crypto.PrepareFileForUpload("text.txt", password)

	// data
	reqBody := models.NoteData{
		EncryptedContent: cipherTextBase64,
		EncryptedKey:     encryptedAESKey,
	}

	// gói data vào json
	jsonBody, _ := json.Marshal(reqBody)

	// tạo request
	req, _ := http.NewRequest("POST", "/notes", bytes.NewBuffer(jsonBody))

	// thiết lập định dạng là json và thêm xác thực
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	if err != nil {
		t.Fatal("Lỗi parse JSON:", err)
	}

	noteId, _ := response["note_id"].(string)

	return noteId
}

// Tạo URL test
func SetupMockURL(t *testing.T, noteId, sender, receiver, expiresIn string, maxAccess int, senderToken string, receiverToken string) string {

	//Tạo URL
	requestBody, _ := json.Marshal(map[string]interface{}{
		"receiver":                 receiver,
		"max_access":               maxAccess,
		"expires_in":               expiresIn,
		"shared_encrypted_aes_key": "mock_shared_key",
		"sender":                   sender,
	})

	reqPost, _ := http.NewRequest("POST", "/notes/"+noteId+"/url", bytes.NewBuffer(requestBody))
	reqPost.Header.Set("Authorization", "Bearer "+senderToken)
	reqPost.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, reqPost)

	if w1.Code != http.StatusOK && w1.Code != 201 {
		t.Fatalf("API tạo URL thất bại, Status: %d, Lỗi: %s", w1.Code, w1.Body.String())
	}

	//Lấy URL
	reqGet, _ := http.NewRequest("GET", "/notes/"+noteId+"/url", nil)
	reqGet.Header.Set("Authorization", "Bearer "+receiverToken)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, reqGet)

	if w2.Code != 200 {
		t.Fatalf("Request thất bại với mã lỗi %d. Body: %s", w2.Code, w2.Body.String())
	}
	//Lấy urlId
	var res struct {
		Url string `json:"url"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&res); err != nil {
		t.Fatalf("Lỗi decode JSON: %v. Body nhận được: %s", err, w2.Body.String())
	}

	// URL format: .../note/{url_id}
	// Cắt chuỗi để lấy phần cuối cùng
	parts := strings.Split(res.Url, "/")
	if len(parts) == 0 {
		t.Fatal("URL trả về rỗng hoặc sai định dạng")
	}
	urlID := parts[len(parts)-1] // Lấy phần tử cuối cùng

	return urlID
}

// Truy cập URL test
func AccessURL(t *testing.T, urlID, recvToken string) int {
	// Gọi API
	targetURL := "/note/" + urlID

	req, _ := http.NewRequest("GET", targetURL, nil)

	//Token để xác thực người nhận
	req.Header.Set("Authorization", "Bearer "+recvToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Trả về Status code để hàm test kiểm tra (200 là OK, 404/403 là lỗi)
	return w.Code
}

// Test
func TestURL(t *testing.T) {
	//Tạo người nhận, người gửi
	senderToken := SetupMockUser(t, "sender_user", "123")
	recvToken := SetupMockUser(t, "receiver_user", "123")

	//KIỂM TRA TRUY CẬP THÀNH CÔNG
	t.Run("Truy cập thành công", func(t1 *testing.T) {
		noteId0 := SetupMockNote(t, "123", senderToken)

		// Tạo link sống 1s
		urlID1 := SetupMockURL(t, noteId0, "sender_user", "receiver_user", "10m", 10, senderToken, recvToken)

		// Truy cập thử
		statusCode := AccessURL(t, urlID1, recvToken)

		assert.Equal(t, http.StatusOK, statusCode, "Truy cập thành công")
	})

	//KIỂM TRA HẾT HẠN THỜI GIAN
	t.Run("Hạn dùng 1 giây", func(t1 *testing.T) {
		noteId1 := SetupMockNote(t, "123", senderToken)

		// Tạo link sống 1s
		urlID1 := SetupMockURL(t, noteId1, "sender_user", "receiver_user", "1s", 10, senderToken, recvToken)

		// Chờ 1 giây để đảm bảo link hết hạn
		fmt.Println("Chờ 1 giây...")
		time.Sleep(1 * time.Second)

		// Truy cập thử
		statusCode := AccessURL(t, urlID1, recvToken)

		// Mong đợi: Server trả về lỗi (404 Not Found hoặc 403 Forbidden) do hết hạn
		assert.NotEqual(t, http.StatusOK, statusCode, "Hết hạn nhưng vẫn truy cập được")
		assert.Equal(t, http.StatusNotFound, statusCode, "Server nên trả về 404 khi link hết hạn")
	})

	//KIỂM TRA HẾT HẠN SỐ LẦN TRUY CẬP
	t.Run("Truy cập tối đa 1 lần", func(t2 *testing.T) {
		noteId2 := SetupMockNote(t, "123", senderToken)

		// Tạo link chỉ được xem 1 lần
		urlID2 := SetupMockURL(t, noteId2, "sender_user", "receiver_user", "1h", 1, senderToken, recvToken)

		// Lần truy cập 1: Phải thành công
		code1 := AccessURL(t, urlID2, recvToken)
		assert.Equal(t, http.StatusOK, code1, "Lần truy cập đầu tiên thành công")

		// Lần truy cập 2: Phải thất bại
		code2 := AccessURL(t, urlID2, recvToken)
		assert.NotEqual(t, http.StatusOK, code2, "Lần truy cập thứ 2 thất bại")
		assert.Equal(t, http.StatusNotFound, code2, "Trạng thái 404 khi link đã hết lượt xem")
	})
}
