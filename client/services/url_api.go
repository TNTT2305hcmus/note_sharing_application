package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"note_sharing_application/client/models"
)

// ---------------------URL------------------------------------------------------
func CreateNoteUrl(noteId, token, sharedEncryptedAESKey, expiresIn, receiver string, maxAccess int, sender string) error {

	// Chuẩn bị dữ liệu (Marshal JSON)
	reqBody := models.Metadata{
		SharedEncryptedAESKey: sharedEncryptedAESKey,
		ExpiresIn:             expiresIn,
		MaxAccess:             maxAccess,
		Receiver:              receiver,
		Sender:                sender,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("lỗi đóng gói JSON: %v", err)
	}

	apiURL := fmt.Sprintf("%s/notes/%s/url", BaseURL, noteId)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("lỗi tạo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("lỗi kết nối server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lỗi từ server (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &result)

	fmt.Println(result.Message)
	return nil
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
