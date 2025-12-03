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
func Register(username, password, pubKeyStr, encryptedPrivkey string) error {
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
		Username:        username,
		Password:        encryptedPassword,
		PublicKey:       pubKeyStr,
		EncrypedPrivKey: encryptedPrivkey,
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

	fmt.Println("Sucesfully Login!\n")
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
