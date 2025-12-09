package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"note_sharing_application/client/crypto"
	"note_sharing_application/client/services"

	"github.com/stretchr/testify/assert"
)

func SetupMockUser(t *testing.T, username, password string) string {
	// Giả lập đăng ký & đăng nhập để lấy token
	mockPubKey := "MockPublicKeyHexString"
	mockEncPrivKey := "MockEncryptedPrivateKeyHex"

	// Lưu ý: Nếu services của bạn chỉ in ra màn hình, bạn nên sửa lại để trả về error/result
	// Ở đây mình giả định services.Register chạy thành công
	services.Register(username, password, mockPubKey, mockEncPrivKey)

	token, _, err := services.Login(username, password)
	if err != nil {
		t.Fatalf("Login failed for %s: %v", username, err)
	}
	return token
}

// Tạo Note test
func SetupMockNote(t *testing.T, password, token string) string {
	cipherTextBase64, encryptedAESKey, err := crypto.PrepareFileForUpload("text.txt", password)
	if err != nil {
		t.Fatalf("Lỗi mã hóa local: %v", err)
	}

	noteID, err := services.CreateNote(token, cipherTextBase64, encryptedAESKey)
	if err != nil {
		t.Fatalf("Tạo note thất bại: %v", err)
	}
	return noteID
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

	reqPost, _ := http.NewRequest("POST", BaseURL+"/notes/"+noteId+"/url", bytes.NewBuffer(requestBody))
	reqPost.Header.Set("Authorization", "Bearer "+senderToken)
	reqPost.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	respPost, err := client.Do(reqPost)
	if err != nil {
		t.Fatalf("Lỗi gọi API tạo URL: %v", err)
	}
	defer respPost.Body.Close()

	if respPost.StatusCode != http.StatusOK && respPost.StatusCode != 201 {
		t.Fatalf("API tạo URL thất bại, Status: %d", respPost.StatusCode)
	}

	//Lấy URL
	reqGet, _ := http.NewRequest("GET", BaseURL+"/notes/"+noteId+"/url", nil)
	reqGet.Header.Set("Authorization", "Bearer "+receiverToken)

	respGet, err := client.Do(reqGet)
	if err != nil {
		t.Fatalf("Lỗi gọi API lấy URL: %v", err)
	}
	defer respGet.Body.Close()

	if respGet.StatusCode != http.StatusOK {
		t.Fatalf("API lấy URL thất bại (Có thể chưa tạo được?), Status: %d", respGet.StatusCode)
	}

	//Lấy urlId
	var res struct {
		Url string `json:"url"`
	}
	if err := json.NewDecoder(respGet.Body).Decode(&res); err != nil {
		t.Fatalf("Lỗi decode JSON từ GET: %v", err)
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
	targetURL := BaseURL + "/note/" + urlID

	req, _ := http.NewRequest("GET", targetURL, nil)

	//Token để xác thực người nhận
	req.Header.Set("Authorization", "Bearer "+recvToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Lỗi kết nối server: %v", err)
		return 0
	}
	defer resp.Body.Close()

	// Trả về Status code để hàm test kiểm tra (200 là OK, 404/403 là lỗi)
	return resp.StatusCode
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
