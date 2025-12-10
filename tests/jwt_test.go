package tests

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	//CHUẨN BỊ DỮ LIỆU TEST
	// Dữ liệu mẫu
	username := "jwt_test"
	password := "jwt_pass"
	// Dữ liệu giả cho DH Key
	mockPubKey := "MockPublicKeyHexString"
	mockEncPrivKey := "MockEncryptedPrivateKeyHex"

	//Đăng ký
	encryptedPass := encryptPasswordForTest(password)
	regBody := map[string]string{
		"username":          username,
		"password":          encryptedPass,
		"public_key":        mockPubKey,
		"encrypted_privKey": mockEncPrivKey,
	}
	regJson, _ := json.Marshal(regBody)

	// Tạo Request giả
	reg, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(regJson))
	reg.Header.Set("Content-Type", "application/json")

	// Ghi lại Response
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reg)

	//Đăng nhập để nhận token về test
	logBody := map[string]string{
		"username": username,
		"password": encryptedPass,
	}
	logJson, _ := json.Marshal(logBody)

	log, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(logJson))
	wLog := httptest.NewRecorder()
	router.ServeHTTP(wLog, log)
	var response map[string]interface{}
	err := json.Unmarshal(wLog.Body.Bytes(), &response)
	if err != nil {
		t.Fatal("Lỗi parse JSON:", err)
	}

	// Lấy ra và ép kiểu
	token, _ := response["token"].(string)

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
			req, _ := http.NewRequest("GET", "/notes/owned", nil)

			if tc.headerValue != "" {
				req.Header.Set("Authorization", tc.headerValue)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "Status Code sai")

			if tc.expectedError != "" {
				assert.Contains(t, w.Body.String(), tc.expectedError, "Thông báo lỗi không khớp")
			}
		})
	}
}
