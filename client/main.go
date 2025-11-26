package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_sharing_application/client/models"
)

// Dung de test dang nhap
func testLoginAPI() {
	data := models.RegisterRequest{
		Username: "QuocThai",
		Password: "1231",
	}

	// struct -> json
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Tao http request
	req, err := http.NewRequest("POST", "http://localhost:8080/api/login", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	// Set header
	req.Header.Set("Content-Type", "application/json")

	// Tao client
	client := &http.Client{}

	// Gui request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// defer dam bao resp.Body.Close() duoc thuc hien truoc khi ra khoi ham
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

}

// Dung de test dang ky
func testRegisterAPI() {
	data := models.RegisterRequest{
		Username:  "QuocThai",
		Password:  "123",
		PublicKey: "456",
	}

	// struct -> json
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Tao http request
	req, err := http.NewRequest("POST", "http://localhost:8080/api/register", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	// Set header
	req.Header.Set("Content-Type", "application/json")

	// Tao client
	client := &http.Client{}

	// Gui request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// defer dam bao resp.Body.Close() duoc thuc hien truoc khi ra khoi ham
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

}

func main() {
	testRegisterAPI()
	testLoginAPI()
}
