package services

import (
	"context"
	"errors"
	"note_sharing_application/server/models"
	"note_sharing_application/server/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShareService struct {
	UrlRepo  *repositories.UrlRepo
	NoteRepo *repositories.NoteRepo
}

func NewShareService(u *repositories.UrlRepo, n *repositories.NoteRepo) *ShareService {
	return &ShareService{UrlRepo: u, NoteRepo: n}
}

// Service: Tạo URL chia sẻ cho một ghi chú
func (s *ShareService) CreateUrl(ctx context.Context, noteIDStr string, expiresIn string, maxAccess int) (string, error) {
	// 1. Convert ID
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return "", errors.New("note ID không hợp lệ")
	}

	// 2. Parse thời gian (vd: "1h", "30m")
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", errors.New("định dạng thời gian không hợp lệ")
	}
	expireTime := time.Now().Add(duration)

	// 3. Tạo Model
	newUrl := models.Url{
		NoteID:    noteID,
		ExpiresAt: expireTime,
		MaxAccess: maxAccess,
	}

	// 4. Gọi Repo lưu
	return s.UrlRepo.InsertOne(ctx, newUrl)
}

func (s *ShareService) GetExistingUrl(ctx context.Context, noteIDStr string) (string, error) {
	// 1. Convert ID
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return "", errors.New("note ID không hợp lệ")
	}

	// 2. Gọi Repo tìm
	urls, err := s.UrlRepo.FindValidUrlsByNoteIDs(ctx, []primitive.ObjectID{noteID})
	if err != nil {
		return "", err // Lỗi DB
	}
	if len(urls) == 0 {
		return "", errors.New("không tìm thấy URL hợp lệ")
	}

	return urls[0].ID.Hex(), nil
}
