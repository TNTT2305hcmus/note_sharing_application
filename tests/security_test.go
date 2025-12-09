package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"note_sharing_application/server/utils"

	"github.com/stretchr/testify/assert"
)

// Test RSA mã hóa và giải mã password
func TestRSA_EncryptionDecryption(t *testing.T) {
	// Sinh server RSA key
	err := utils.GenerateServerRSAKeys()
	assert.NoError(t, err, "Server phải sinh được khóa RSA")
	assert.NotNil(t, utils.ServerPublicKey, "Public Key không được nil")
	assert.NotNil(t, utils.ServerPrivateKey, "Private Key không được nil")

	// Dữ liệu mẫu
	originalPassword := "pass"

	// Client giả mã hóa password bằng Server RSA Pubkey
	rng := rand.Reader
	// OAEP label
	label := []byte("")

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rng,
		utils.ServerPublicKey,
		[]byte(originalPassword),
		label,
	)
	assert.NoError(t, err, "Client mã hóa thất bại")

	// Server giải mã password
	decryptedString, err := utils.DecryptOAEP(encryptedBytes)

	// Check
	assert.NoError(t, err, "Server giải mã thất bại")
	assert.Equal(t, originalPassword, decryptedString, "Password sau khi giải mã phải khớp với bản gốc")
}

// Test server hash password khi lưu vào db
func TestHashPasswordBySalt(t *testing.T) {
	rawPassword := "pass"
	wrongPassword := "WrongPass"

	// Sinh salt ngẫu nheien
	salt, err := utils.GenerateSalt()
	assert.NoError(t, err)
	assert.Len(t, salt, 32, "Salt (Hex) thường dài 32 ký tự cho 16 bytes")

	// Hash pass bằng salt
	hashedPassword := utils.HashPassword(rawPassword, salt)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, rawPassword, hashedPassword, "Password phải được băm, không được lưu plain text")

	// Check pass đúng
	match := utils.CheckPasswordHash(rawPassword, salt, hashedPassword)
	assert.True(t, match, "CheckPasswordHash phải trả về TRUE khi nhập đúng pass")

	// Check pass sai
	matchWrong := utils.CheckPasswordHash(wrongPassword, salt, hashedPassword)
	assert.False(t, matchWrong, "CheckPasswordHash phải trả về FALSE khi nhập sai pass")

	// Check salt
	// (Cùng pass, khác salt => Khác hash)
	salt2, _ := utils.GenerateSalt()
	hashedPassword2 := utils.HashPassword(rawPassword, salt2)
	assert.NotEqual(t, hashedPassword, hashedPassword2, "Cùng password nhưng khác Salt thì Hash phải khác nhau")
}
