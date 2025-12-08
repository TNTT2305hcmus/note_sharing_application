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

var BaseURL = "http://localhost:8080"

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
		return fmt.Errorf("Lỗi: Không thể mã hóa mật khẩu bằng Server Public Key RSA: %v", err)
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
		return fmt.Errorf("Lỗi: Đăng ký không thành công: %s", string(body))
	}
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
		return "", "", fmt.Errorf("Lỗi: Không thể mã hóa mật khẩu bằng Server Public Key RSA: %v", err)
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
		return "", "", fmt.Errorf("Đăng nhập không thành công: %s", result.Error)
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
