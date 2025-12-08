package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Tạo khóa đối xứng từ Password (sử dụng PBKDF2)
// Salt nên được sinh ngẫu nhiên mỗi lần tạo user
// Iterations: 100,000 lần, KeyLen: 32 bytes (AES-256)
func DeriveKeyFromPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}

// Hàm mã hóa Private Key bằng Password
// Output format: Salt (16 bytes) + Nonce (12 bytes) + CipherText
func EncryptByPassword(privKeyHex string, password string) (string, error) {
	// Tạo Salt ngẫu nhiên cho việc dẫn xuất khóa
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Tạo khóa đối xứng từ password + salt vừa tạo
	key := DeriveKeyFromPassword(password, salt)

	// Tạo AES Block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Sử dụng chế độ GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Tạo Nonce ngẫu nhiên
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Mã hóa Private Key
	ciphertext := aesGCM.Seal(nil, nonce, []byte(privKeyHex), nil)

	// Đóng gói kết quả: Salt + Nonce + CipherText
	finalData := append(salt, nonce...)
	finalData = append(finalData, ciphertext...)

	return hex.EncodeToString(finalData), nil
}

// 3. Hàm giải mã Private Key khi đăng nhập
func DecryptByPassword(encryptedHex string, password string) (string, error) {
	data, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	if len(data) < 16+12 {
		return "", errors.New("dữ liệu không hợp lệ")
	}

	// Tách Salt, Nonce và CipherText
	salt := data[:16]
	nonce := data[16 : 16+12]
	ciphertext := data[16+12:]

	// Tính lại khóa đối xứng từ password + salt cũ
	key := DeriveKeyFromPassword(password, salt)

	// Giải mã
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("error password or wrong data")
	}

	return string(plaintext), nil
}
