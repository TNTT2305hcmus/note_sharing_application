package tests

import (
	"crypto/rand"
	"testing"

	"note_sharing_application/client/crypto"

	"github.com/stretchr/testify/assert"
)

// Tạo dữ liệu ngẫu nhiên để test
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// Test mã hóa PrivKey và AESKey bằng password
func TestKeyProtectionByPassword(t *testing.T) {
	// Dữ liệu mẫu
	password := "pass"
	wrongPassword := "WrongPass"

	// Giả 1 khóa PrivKey hoặc AES Key dạng string (Hex)
	originalKeyHex := "abcdef1234567890abcdef1234567890"

	t.Run("Encrypt and Decrypt with Correct Password", func(t *testing.T) {
		// Mã hóa key bằng password
		encryptedKey, err := crypto.EncryptByPassword(originalKeyHex, password)
		assert.NoError(t, err)
		assert.NotEqual(t, originalKeyHex, encryptedKey)

		// Giải mã key lại bằng password
		decryptedKey, err := crypto.DecryptByPassword(encryptedKey, password)
		assert.NoError(t, err)

		// Check đúng
		assert.Equal(t, originalKeyHex, decryptedKey, "Khóa giải mã được phải trùng khớp khóa gốc")
	})

	t.Run("Decrypt with Wrong Password", func(t *testing.T) {
		// Mã hóa key bằng password
		encryptedKey, _ := crypto.EncryptByPassword(originalKeyHex, password)

		// Giải mã bằng wrong pass
		_, err := crypto.DecryptByPassword(encryptedKey, wrongPassword)

		// Đợi lỗi
		assert.Error(t, err, "Phải báo lỗi khi nhập sai mật khẩu")
	})

}

// Test mã hóa bằng AES Key
func TestAESGCM_EncryptionDecryption(t *testing.T) {
	// Sinh khóa AES ngẫu nhiên
	key, err := crypto.GenerateAESKey()
	assert.NoError(t, err, "Lỗi sinh khóa AES")
	assert.Len(t, key, 32, "Khóa AES phải dài 32 bytes (256-bit)")

	// Dữ liệu giả lập (Nội dung ghi chú)
	originalText := []byte("Nội dung ghi chú")

	// Mã hóa bằng AES Key
	ciphertext, err := crypto.EncryptBytes(originalText, key)
	assert.NoError(t, err, "Mã hóa thất bại")
	assert.NotEqual(t, originalText, ciphertext, "Ciphertext không được giống Plaintext")

	// Giải mã bằng AES Key
	plaintext, err := crypto.DecryptBytes(ciphertext, key)
	assert.NoError(t, err, "Giải mã thất bại")

	// Check dữ liệu
	assert.Equal(t, originalText, plaintext, "Dữ liệu sau giải mã phải khớp 100 với gốc")
}

// Test Diffie-Hellman
func TestSharedKeyExchange(t *testing.T) {
	// Giải sử 2 người dùng (Alice - Bob)

	// Sinh khóa DH cho alice
	alicePriv, alicePub, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)
	alicePubHex := alicePub.Text(16)

	// Sinh khóa DH cho bob
	bobPriv, bobPub, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)
	bobPubHex := bobPub.Text(16)

	// alice tính K chung
	aliceSharedK, err := crypto.ComputeSharedSecret(alicePriv, bobPubHex)
	assert.NoError(t, err)

	// bob tính K chung
	bobSharedK, err := crypto.ComputeSharedSecret(bobPriv, alicePubHex)
	assert.NoError(t, err)

	// Check khóa K
	assert.Equal(t, aliceSharedK.Text(16), bobSharedK.Text(16), "Khóa chung Diffie-Hellman không khớp!")

	// Test quá trình chia sẻ AES thông qua khóa phiên K
	t.Run("Wrap and Unwrap AES Key with Shared K", func(t *testing.T) {
		// Tạo khóa AES mẫu
		originalAESKey := generateRandomBytes(32) // 32 bytes

		// alice mã AES key bằng aliceSharedK
		encryptedKeyForBob, err := crypto.EncryptAESKeyWithSharedK(originalAESKey, aliceSharedK)
		assert.NoError(t, err)

		// bob giải mã encryptedKeyForBob bằng bobSharedK
		decryptedAESKey, err := crypto.DecryptAESKeyWithSharedK(encryptedKeyForBob, bobSharedK)
		assert.NoError(t, err)

		// Check khóa AES
		assert.Equal(t, originalAESKey, decryptedAESKey, "Khóa AES sau khi chia sẻ bị sai lệch")
	})
}
