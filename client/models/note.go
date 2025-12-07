package models

import "time"

type Note struct {
	ID               int       `json:"id"`
	OwnerID          int       `json:"owner_id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	CreatedAt        time.Time `json:"created_at"`
}

type NoteData struct {
	EncryptedContent string `json:"cipher_text"`
	EncryptedKey     string `json:"encrypted_aes_key"`
}
