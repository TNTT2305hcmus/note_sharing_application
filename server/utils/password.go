package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

// GenerateSalt tạo 1 random string (16 bytes)
func GenerateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword tạo mã băm từ password và salt sử dụng PBKDF2
// Iterations: 100000 (đủ an toàn), KeyLen: 32 bytes (cho SHA-256)
func HashPassword(password string, salt string) string {
	// Chuyển salt từ Hex string sang byte
	saltBytes, _ := hex.DecodeString(salt)

	// Thực hiện thuật toán PBKDF2
	dk := pbkdf2.Key([]byte(password), saltBytes, 100000, 32, sha256.New)

	// Trả về chuỗi base64 để lưu vào DB
	return base64.StdEncoding.EncodeToString(dk)
}

func CheckPasswordHash(password, salt, sortedPassword string) bool {
	newHash := HashPassword(password, salt)
	return newHash == sortedPassword
}
