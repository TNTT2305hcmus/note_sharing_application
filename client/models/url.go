package models

import (
	"time"
)

type Metadata struct {
	ExpiresIn string `json:"expires_in"` // "1h", "30m"
	MaxAccess int    `json:"max_access"` // int (Server yêu cầu số)
}

// Url đại diện cho thông tin đường dẫn chia sẻ
type Url struct {
	ID         string    `json:"url_id"` // Khớp với json tag của ObjectID bên server
	NoteID     string    `json:"note_id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	MaxAccess  int       `json:"max_access"`
}
