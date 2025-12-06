package services

import (
	"context"
	"errors"
	"note_sharing_application/server/models"
	"note_sharing_application/server/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteService struct {
	NoteRepo *repositories.NoteRepo
	UrlRepo  *repositories.UrlRepo
}

func NewNoteService(n *repositories.NoteRepo, u *repositories.UrlRepo) *NoteService {
	return &NoteService{NoteRepo: n, UrlRepo: u}
}

// Service: xem tất cả ghi chú do một owner
func (s *NoteService) ViewOwnedNotes(ctx context.Context, ownerIDStr string) ([]models.Note, error) {
	ownerID, err := primitive.ObjectIDFromHex(ownerIDStr)
	if err != nil {
		return nil, errors.New("User ID không hợp lệ")
	}
	return s.NoteRepo.FindByOwnerID(ctx, ownerID)
}

// Servce: xem tất cả ghi chú được gửi đến receiver
func (s *NoteService) ViewReceivedNotes(ctx context.Context, receiverIDStr string) ([]models.Note, error) {

	receiverID, err := primitive.ObjectIDFromHex(receiverIDStr)
	if err != nil {
		return nil, errors.New("User ID không hợp lệ")
	}

	// tìm tất cả note được gửi đến receiver
	allReceivedNotes, err := s.NoteRepo.FindByReceiverID(ctx, receiverID)
	if err != nil {
		return nil, err
	}

	// nếu không có note nào, trả về rỗng ngay
	if len(allReceivedNotes) == 0 {
		return []models.Note{}, nil
	}

	//!! lấy danh sách ID của các note trên
	var noteIDs []primitive.ObjectID
	noteMap := make(map[primitive.ObjectID]models.Note)

	for _, note := range allReceivedNotes {
		noteIDs = append(noteIDs, note.ID)
		noteMap[note.ID] = note
	}

	// lọc ra các URL hợp lệ tương ứng với các noteID trên
	validUrls, err := s.UrlRepo.FindValidUrlsByNoteIDs(ctx, noteIDs)
	if err != nil {
		return nil, err
	}

	// lưu kết quả cuối cùng
	var finalNotes []models.Note
	for _, url := range validUrls {
		if note, exists := noteMap[url.NoteID]; exists {
			finalNotes = append(finalNotes, note)
		}
	}

	// trả về kết quả
	return finalNotes, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteIDStr string, requesterIDStr string) error {

	// string -> ObjectID
	noteID, err := primitive.ObjectIDFromHex(noteIDStr)
	if err != nil {
		return errors.New("Note ID không hợp lệ")
	}

	// lấy note từ DB
	note, err := s.NoteRepo.FindByID(ctx, noteID)
	if err != nil {
		return errors.New("ghi chú không tồn tại")
	}

	// kiểm tra quyền sở hữu
	if note.OwnerID.Hex() != requesterIDStr {
		return errors.New("bạn không có quyền xóa ghi chú này") // Lỗi 403 Forbidden
	}

	// xóa trong collection note
	if err := s.NoteRepo.DeleteByID(ctx, noteID); err != nil {
		return err
	}

	// xóa trong collection url
	_ = s.UrlRepo.DeleteByNoteID(ctx, noteID)

	return nil

}

