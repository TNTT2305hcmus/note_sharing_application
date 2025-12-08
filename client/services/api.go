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

const BaseURL = "http://localhost:8080"

// Struct nhận server public key RSA
type PublicKeyResponse struct {
	ServerPublicKeyRSA string `json:"server-public-key-rsa"`
}

// Struct nhận client public key
type UserPublicKeyResponse struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

// --------------------- AUTH GROUP ---------------------
// URL = BaseURL + /auth

func GetServerPublicKeyRSA() (string, error) {
	url := BaseURL + "/auth/server-public-key-rsa"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Cant get server public key rsa: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error get server public key rsa: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var keyRes PublicKeyResponse
	if err := json.Unmarshal(body, &keyRes); err != nil {
		return "", fmt.Errorf("Error read JSON: %v", err)
	}

	return keyRes.ServerPublicKeyRSA, nil
}

func Register(username, password, pubKeyStr, EncryptedPrivateKey string) error {
	serverRSAPubKey, err := GetServerPublicKeyRSA()
	if err != nil {
		return err
	}

	// Mã hóa password thô bằng RSA để gửi lên Server qua http
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

	resp, err := http.Post(BaseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Unsuccessfully Registered: %s", string(body))
	}

	fmt.Println("Successfully Registered!")
	return nil
}

func Login(username, password string) (string, string, error) {
	serverRSAPubKey, err := GetServerPublicKeyRSA()
	if err != nil {
		return "", "", err
	}

	// Mã hóa password thô bằng RSA để gửi lên server thông qua http
	encryptedPassword, err := crypto.EncryptPasswordWithServerKey(password, serverRSAPubKey)
	if err != nil {
		return "", "", fmt.Errorf("Error encrypted password: %v", err)
	}

	reqBody := models.LoginRequest{
		Username: username,
		Password: encryptedPassword,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(BaseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result models.LoginResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Unsuccessfully Login: %s", result.Error)
	}
	return result.Token, result.EncryptedPrivateKey, nil
}

func GetUserPublicKey(targetUsername string) (string, error) {
	url := fmt.Sprintf("%s/auth/users/%s/pubkey", BaseURL, targetUsername)

	resp, err := http.Get(url)
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

// --------------------- NOTE GROUP ---------------------
// URL = BaseURL + /notes

func CreateNoteUrl(noteId, token, expiresIn string, maxAccess int) (string, error) {
	reqBody := models.Metadata{
		ExpiresIn: expiresIn,
		MaxAccess: maxAccess,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("lỗi đóng gói JSON: %v", err)
	}

	apiURL := fmt.Sprintf("%s/notes/%s/url", BaseURL, noteId)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("lỗi tạo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

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
		return "", fmt.Errorf("lỗi đọc kết quả: %v", err)
	}
	return result.Url, nil
}

func GetNoteUrl(noteId, token string) (string, error) {
	url := fmt.Sprintf("%s/notes/%s/url", BaseURL, noteId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

func ReadNoteWithURL(urlId, token string) (models.NoteData, error) {
	var result models.NoteData

	// Lưu ý: urlId ở đây là ID của URL chia sẻ (ObjectID hex)
	url := fmt.Sprintf("%s/note/%s", BaseURL, urlId)

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

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("lỗi cấu trúc JSON: %v", err)
	}

	return result, nil
}
