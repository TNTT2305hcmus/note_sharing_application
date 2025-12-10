package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"note_sharing_application/client/services"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	BaseURL = "http://localhost:8080"
)

func TestJWT(t *testing.T) {
	client := &http.Client{}

	//CHUẨN BỊ DỮ LIỆU TEST
	// Dữ liệu mẫu
	username := "jwt_test"
	password := "jwt_pass"
	// Dữ liệu giả cho DH Key
	mockPubKey := "MockPublicKeyHexString"
	mockEncPrivKey := "MockEncryptedPrivateKeyHex"

	services.Register(username, password, mockPubKey, mockEncPrivKey)

	//Đăng nhập để nhận token về test
	token, _, _ := services.Login(username, password)
	//Nhận token để test

	// Token bị sửa body
	testBodyToken := token[:len(token)-5] + "XXXXX"

	//Token bị sửa signature
	parts := strings.Split(token, ".")
	signingInput := parts[0] + "." + parts[1]

	h := hmac.New(sha256.New, []byte("Fake sig"))
	h.Write([]byte(signingInput))

	fakeSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	testSigToken := signingInput + "." + fakeSig

	// Token bị sửa header
	noneHeader := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0"
	testHeaderToken := noneHeader + "." + parts[1] + "." + parts[2]

	//TEST CASES
	testCases := []struct {
		name           string
		headerValue    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Case 0: Token hợp lệ",
			headerValue:    "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Case 1: Thiếu Authorization Header",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token không hợp lệ",
		},
		{
			name:           "Case 2: Thiếu chữ Bearer",
			headerValue:    token,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Định dạng xác thực sai",
		},
		{
			name:           "Case 3: Chỉ có chữ Bearer",
			headerValue:    "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Định dạng xác thực sai",
		},
		{
			name:           "Case 4: Token bị sửa đổi body",
			headerValue:    "Bearer " + testBodyToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token không hợp lệ hoặc đã hết hạn",
		},
		{
			name:           "Case 5: Token sửa header",
			headerValue:    "Bearer " + testHeaderToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token không hợp lệ hoặc đã hết hạn",
		},
		{
			name:           "Case 6: Token sửa signature",
			headerValue:    "Bearer " + testSigToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token không hợp lệ hoặc đã hết hạn",
		},
	}

	//TEST
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", BaseURL+"/notes/owned", nil)

			if tc.headerValue != "" {
				req.Header.Set("Authorization", tc.headerValue)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Lỗi gọi API: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Status Code sai")

			if tc.expectedError != "" {
				assert.Contains(t, bodyString, tc.expectedError, "Thông báo lỗi không khớp")
			}
		})
	}
}
