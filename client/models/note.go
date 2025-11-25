package models

import "time"

type Note struct {
	ID               string    `json:"id"`
	OwnerID          string    `json:"owner_id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	CreatedAt        time.Time `json:"created_at"`
}
