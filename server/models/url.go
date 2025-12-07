package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Url struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"url_id"`
	NoteID    string `bson:"note_id" json:"note_id"`       // ID của ghi chú gốc
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"` // Thời gian hết hạn
	MaxAccess int                `bson:"max_access" json:"max_access"` // Số lượt truy cập tối đa
	Accessed  int                `bson:"accessed" json:"accessed"`     // Số lượt đã truy cập

}

type CreateUrlRequest struct {
	ExpiresIn string `json:"expires_in"` // vd: "1h"
	MaxAccess int    `json:"max_access"` // vd: 5
}
