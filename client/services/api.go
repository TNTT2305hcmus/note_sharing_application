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
