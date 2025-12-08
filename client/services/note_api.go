package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_sharing_application/client/models"
)

// --------------------- NOTE GROUP ---------------------
// URL = BaseURL + /notes

// Tạo một note
func CreateNote(token, cipherText, encryptedAESKey string) (string, error) {

	// url
	apiURL := fmt.Sprintf("%s/notes", BaseURL)

	// data
	reqBody := models.NoteData{
		EncryptedContent: cipherText,
		EncryptedKey:     encryptedAESKey,
	}

	// gói data vào json
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("lỗi đóng gói JSON: %v", err)
	}

	// tạo request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("lỗi tạo request: %v", err)
	}

	// thiết lập định dạng là json và thêm xác thực
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Client giúp tạo kết nối TCP đến Server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	//? Lấy body?
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Chấp nhận cả 200 (OK) và 201 (Created)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	// Nhận lại noteID đã được tạo ở server
	var result struct {
		// Viết hoa để có thể truy cập
		Note_id string `json:"note_id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("lỗi đọc kết quả: %v", err)
	}
	return result.Note_id, nil
}

// xóa một note
func DeleteNote(token, noteID string) error {

	// tạo URL
	apiURL := fmt.Sprintf("%s/notes/%s", BaseURL, noteID)

	// tạo request với method DELETE
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("lỗi tạo request: %v", err)
	}

	// thiết lập Header xác thực
	req.Header.Set("Authorization", "Bearer "+token)

	// gửi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	// đọc phản hồi từ Server
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lỗi đọc phản hồi: %v", err)
	}

	// check status code
	if resp.StatusCode != http.StatusOK {

		// hứng lỗi từ server
		var errResponse struct {
			Error string `json:"error"`
		}

		// Cố gắng parse JSON lỗi
		if jsonErr := json.Unmarshal(body, &errResponse); jsonErr == nil && errResponse.Error != "" {
			return fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, errResponse.Error)
		}

		// Nếu server không trả JSON chuẩn, in raw body
		return fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}
	// Xóa thành công
	return nil
}

// lấy danh sách tất cả ghi chú của người dùng hiện tại
func GetOwnedNotes(token string) ([]models.Note, error) {

	// tạo URL
	apiURL := fmt.Sprintf("%s/notes/owned", BaseURL)

	// tạo request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo request: %v", err)
	}

	// gắn Header xác thực
	req.Header.Set("Authorization", "Bearer "+token)

	// gửi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	// đọc phản hồi
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc dữ liệu phản hồi: %v", err)
	}

	// kiểm tra Status Code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	// giải mã JSON
	var notes []models.Note
	if err := json.Unmarshal(body, &notes); err != nil {
		return nil, fmt.Errorf("lỗi giải mã JSON: %v", err)
	}

	return notes, nil
}

// lấy danh sách cá URLs được chia sẽ
func GetReceivedURLs(token string) ([]models.Url, error) {

	// tạo url
	apiURL := fmt.Sprintf("%s/notes/received", BaseURL)

	// tạo Request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo request: %v", err)
	}

	// gắn Header xác thực
	req.Header.Set("Authorization", "Bearer "+token)

	// gửi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	// đọc body phản hồi
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc phản hồi: %v", err)
	}

	// kiểm tra mã lỗi
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	// giải mã JSON
	var urls []models.Url
	if err := json.Unmarshal(body, &urls); err != nil {
		return nil, fmt.Errorf("lỗi giải mã JSON: %v", err)
	}

	return urls, nil
}

// xóa chia sẻ note
func DeleteSharedNote(token, noteID string) error {

	// tạo URL
	apiURL := fmt.Sprintf("%s/notes/shared/%s", BaseURL, noteID)

	// tạo Request
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("lỗi tạo request: %v", err)
	}

	// gắn Header xác thực
	req.Header.Set("Authorization", "Bearer "+token)

	// gửi Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	// đọc body (để lấy thông báo lỗi nếu có)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("lỗi đọc phản hồi: %v", err)
	}

	// kiểm tra Status Code
	if resp.StatusCode != http.StatusOK {
		// Cấu trúc hứng lỗi JSON
		var errResponse struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errResponse); jsonErr == nil && errResponse.Error != "" {
			return fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, errResponse.Error)
		}
		return fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}
