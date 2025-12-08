package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

func deriveKeyFromK(sharedK *big.Int) []byte {
	// Chuyển BigInt thành mảng byte
	kBytes := sharedK.Bytes()

	// Hash SHA-256 để luôn đảm bảo đầu ra là 32 bytes (cho AES-256)
	hash := sha256.Sum256(kBytes)
	return hash[:]
}

// Mã hóa AES key bằng khóa phiên K
// Output:
//   - String Hex: Dạng "Nonce + Ciphertext"
func EncryptAESKeyWithSharedK(aesKeyRaw []byte, sharedK *big.Int) (string, error) {
	// Tạo wrappingKey từ khóa phiên K chung
	wrappingKey := deriveKeyFromK(sharedK)

	// Tạo Block Cipher
	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return "", fmt.Errorf("lỗi tạo cipher từ K: %v", err)
	}

	// Dùng mode GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("lỗi tạo GCM: %v", err)
	}

	// Tạo nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("lỗi sinh nonce: %v", err)
	}

	// Mã hóa (Seal)
	// Kết quả = Nonce  + dataEncrypted
	encryptedBytes := gcm.Seal(nonce, nonce, aesKeyRaw, nil)

	return hex.EncodeToString(encryptedBytes), nil
}

// Giải mã EncryptedAESKey bằng K
// Output:
//   - []byte: Khóa AES gốc dùng để giải mã ghi chú
func DecryptAESKeyWithSharedK(encryptedHex string, sharedK *big.Int) ([]byte, error) {
	// Decode Hex sang byte
	data, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, fmt.Errorf("chuỗi mã hóa không phải hex hợp lệ: %v", err)
	}

	// Tạo wrapping key từ K
	wrappingKey := deriveKeyFromK(sharedK)

	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Tách nonce và ciphertext(dataEncrypted)
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("dữ liệu quá ngắn, lỗi định dạng")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypted
	plaintextAESKey, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("giải mã thất bại (K sai hoặc dữ liệu bị sửa đổi): %v", err)
	}

	return plaintextAESKey, nil
}
