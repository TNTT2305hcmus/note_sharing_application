package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_sharing_application/client/crypto"
	"note_sharing_application/client/models"
)

// Server
const BaseURL = "http://localhost:8080/api"

// Struct nhận server public key RSA nhằm mã hóa password trước khi gửi cho server
type PublicKeyResponse struct {
	ServerPublicKeyRSA string `json:"server-public-key-rsa"`
}

// Struct nhận client public key để tính khóa phiên K
type UserPublicKeyResponse struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

// Gọi API trả về server-public-key-rsa
func GetServerPublicKeyRSA() (string, error) {
	// Gọi API Lấy server-public-key-rsa
	resp, err := http.Get(BaseURL + "/server-public-key-rsa")
	if err != nil {
		return "", fmt.Errorf("Cant get server public key rsa: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error get server public key rsa: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	// Parse JSON
	var keyRes PublicKeyResponse
	if err := json.Unmarshal(body, &keyRes); err != nil {
		return "", fmt.Errorf("Error read JSON get server public key rsa: %v", err)
	}

	return keyRes.ServerPublicKeyRSA, nil
}

// Yêu cầu đăng ký tài khoản
func Register(username, password, pubKeyStr, EncryptedPrivateKey string) error {
	// Lấy Server public key RSA để mã hóa password
	serverRSAPubKey, err := GetServerPublicKeyRSA()
	if err != nil {
		return err
	}

	// Mã hóa password
	encryptedPassword, err := crypto.EncryptPasswordWithServerKey(password, serverRSAPubKey)
	if err != nil {
		return fmt.Errorf("Error encrypted password: %v", err)
	}

	reqBody := models.RegisterRequest{
		Username:            username,
		Password:            encryptedPassword,
		PublicKey:           pubKeyStr,
		EncryptedPrivateKey: EncryptedPrivateKey,
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

// Yêu cầu đăng nhập
func Login(username, password string) (string, string, error) {
	// Lấy Server public key RSA để mã hóa password
	serverRSAPubKey, err := GetServerPublicKeyRSA()
	if err != nil {
		return "", "", err
	}

	// Mã hóa password
	encryptedPassword, err := crypto.EncryptPasswordWithServerKey(password, serverRSAPubKey)
	if err != nil {
		return "", "", fmt.Errorf("Error encrypted password: %v", err)
	}

	reqBody := models.LoginRequest{
		Username: username,
		Password: encryptedPassword,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(BaseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result models.LoginResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Unsucesfully Login: %s", result.Error)
	}
	return result.Token, result.EncryptedPrivateKey, nil
}

// Gọi API trả về public_key của đối phương
func GetUserPublicKey(targetUsername string) (string, error) {
	resp, err := http.Get(BaseURL + "/users/" + targetUsername + "/pubkey")
	if err != nil {
		return "", fmt.Errorf("Error connected: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", fmt.Errorf("User '%s' is not exist", targetUsername)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error Server: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var res UserPublicKeyResponse
	json.Unmarshal(body, &res)

	return res.PublicKey, nil
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

// ------------------------------------Notes-------------------------------------------