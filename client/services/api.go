package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_sharing_application/client/models"
)

// Server
const BaseURL = "http://localhost:8080/api"

func Register(username, password, pubKeyStr string) error {
	reqBody := models.RegisterRequest{
		Username:  username,
		Password:  password,
		PublicKey: pubKeyStr,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(BaseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Unsucesfully Registed: %s", string(body))
	}

	fmt.Println("Sucesfully Registed!")
	return nil
}

func Login(username, password string) (string, error) {
	reqBody := models.LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(BaseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result models.LoginResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unsucesfully Login: %s", result.Error)
	}

	fmt.Println("Sucesfully Login! - Token:", result.Token)
	return result.Token, nil
}

// ---------------------URL------------------------------------------------------
func CreateNoteUrl(noteId, token, expiresIn string, maxAccess int) (string, error) {

	// Chuẩn bị dữ liệu (Marshal JSON)
	reqBody := models.Metadata{
		ExpiresIn: expiresIn,
		MaxAccess: maxAccess,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("lỗi đóng gói JSON: %v", err)
	}

	apiURL := fmt.Sprintf(BaseURL+"/notes/%s/url", noteId)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("lỗi tạo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	//  Gửi Request
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	// Đọc Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Kiểm tra Status Code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Url string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("lỗi đọc kết quả: %v", err)
	}
	return result.Url, nil
}

func GetNoteUrl(noteId, token string) (string, error) {
	url := fmt.Sprintf("%s/notes/%s/url", BaseURL, noteId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Thực thi
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Đọc Body
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Url string `json:"url"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("lỗi đọc JSON: %v", err)
	}

	return result.Url, nil
}

func ReadNoteWithURL(url, token string) (models.NoteData, error) {
	var result models.NoteData

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return result, fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("lỗi đọc dữ liệu: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	// 5. Parse JSON thành công vào struct NoteData
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("lỗi cấu trúc JSON: %v", err)
	}

	return result, nil
}
