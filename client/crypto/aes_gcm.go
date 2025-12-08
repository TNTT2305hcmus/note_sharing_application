package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// hàm sinh AES Key
func GenerateAESKey() ([]byte, error) {

	// khởi tạo mảng byte có kích thước 32 (256 bits)
	key := make([]byte, 32)

	// sinh AES Key
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("không thể tạo AES Key: %w", err)
	}

	return key, nil
}

// Hàm mã hóa file
func EncryptFile(inputFile string, outputFile string, key []byte) error {
	// đọc toàn bộ input file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("lỗi đọc file đầu vào: %w", err)
	}

	// tạo block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("lỗi tạo block cipher: %w", err)
	}

	// aesGCM hỗ trợ mã hóa
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("lỗi tạo GCM: %w", err)
	}

	// tạo nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("lỗi tạo nonce: %w", err)
	}

	// thực hiện mã hóa
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// trả kết quả
	err = os.WriteFile(outputFile, ciphertext, 0644)
	if err != nil {
		return fmt.Errorf("lỗi ghi file đầu ra: %w", err)
	}

	return nil

}

// hàm giải mã file
func DecryptedByAESKey(key []byte, inputFile string, outputFile string) error {
	// đọc toàn bộ input file
	ciphertext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("lỗi đọc file mã hóa: %w", err)
	}

	// tạo block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("lỗi tạo block cipher: %w", err)
	}

	// aesGCM hỗ trợ mã hóa
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("lỗi tạo GCM: %w", err)
	}

	// kiểm tra kích thước file có hợp lệ không
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("dữ liệu quá ngắn, không đúng định dạng")
	}

	// tách nonce
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// giải mã
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return fmt.Errorf("giải mã thất bại (sai khóa hoặc file bị sửa đổi): %w", err)
	}

	// trả kết quả
	err = os.WriteFile(outputFile, plaintext, 0644)
	if err != nil {
		return fmt.Errorf("lỗi ghi file giải mã: %w", err)
	}

	return nil
}

// ! Lưu ý: cho cipher text và aes key được biểu diễn ở dạng binary nên mình sẽ code 2 hàm dưới đây để chuyển byte[] -> string (base64)
// hàm chuyển binary -> string
func ConvertBinaryToString(encryptedFilePath string) (string, error) {
	// đọc dữ liệu binary từ file
	fileData, err := os.ReadFile(encryptedFilePath)
	if err != nil {
		return "", err
	}

	// mã hóa sang Base64
	// Base64 biến đổi binary thành các ký tự an toàn: A-Z, a-z, 0-9, +, /
	encodedString := base64.StdEncoding.EncodeToString(fileData)

	return encodedString, nil
}

// hàm chuyển string -> byte[]
func ConvertStringToBinary(base64String string, outputPath string) error {
	// giải mã từ Base64 về lại binary gốc
	decodedData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Errorf("dữ liệu base64 bị lỗi: %w", err)
	}

	// viết ra output
	err = os.WriteFile(outputPath, decodedData, 0644)
	if err != nil {
		return err
	}
	return nil
}
