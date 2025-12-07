package handlers

import (
	"fmt"
	"net/http"
	"note_sharing_application/server/models"
	"note_sharing_application/server/services"

	"github.com/gin-gonic/gin"
)

// Tạo URL (POST /api/:note_id/url)
func CreateNoteUrl(c *gin.Context) {
	noteId := c.Param("note_id")
	expiresIn := c.GetString("expires_in")
	maxAccess := c.GetInt("max_access")

	// Gọi Service tạo đối tượng trong DB
	urlId, err := services.CreateUrl(noteId, expiresIn, maxAccess)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trả về đường dẫn
	finalUrl := fmt.Sprintf("localhost:8080/view/%s", urlId)
	c.JSON(http.StatusOK, gin.H{"url": finalUrl})
}

// GET /api/:note_id/url
func GetNoteUrl(c *gin.Context) {
	noteId := c.Param("note_id")

	// Gọi Service xem DB có URL nào không
	urlId, err := services.GetExistingUrl(noteId)

	if err != nil {
		// Không có thì gửi message báo lỗi
		c.JSON(http.StatusNotFound, gin.H{"message": "Chưa có URL chia sẻ nào cho note này"})
		return
	}

	// Còn thì gửi về url
	finalUrl := fmt.Sprintf("https:://localhost:8080/note/%s", urlId)
	c.JSON(http.StatusOK, gin.H{"url": finalUrl})
}

// (GET note/:url_id)
func ViewNoteHandler(c *gin.Context) {
	// Client đã lấy url_id và gọi API này
	url := c.MustGet("url").(models.Url)

	note, err := services.GetNote(url)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Trả về {cipher_text, encrypted_aes_key}
	c.JSON(http.StatusOK, gin.H{
		"cipher_text":       note.CipherText,
		"encrypted_aes_key": note.EncryptedAesKey,
	})
}
