package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
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

// Mã file
func PrepareFileForUpload(filePath string, password string) (string, string, error) {

	// sinh AES Key ngẫu nhiên (32 bytes)
	aesKey, err := GenerateAESKey()
	if err != nil {
		return "", "", fmt.Errorf("lỗi sinh khóa AES: %v", err)
	}

	// mã hóa File
	tempEncryptedPath := filePath + ".enc_temp"

	// đảm bảo xóa file tạm này khi hàm chạy xong
	defer os.Remove(tempEncryptedPath)

	err = EncryptFile(filePath, tempEncryptedPath, aesKey)
	if err != nil {
		return "", "", fmt.Errorf("lỗi mã hóa file: %v", err)
	}

	// chuyển File đã mã hóa sang Base64
	cipherTextBase64, err := ConvertBinaryToString(tempEncryptedPath)
	if err != nil {
		return "", "", fmt.Errorf("lỗi chuyển đổi file sang base64: %v", err)
	}

	// mã hóa AES Key bằng Password
	aesKeyHex := hex.EncodeToString(aesKey)

	encryptedAESKey, err := EncryptByPassword(aesKeyHex, password)
	if err != nil {
		return "", "", fmt.Errorf("lỗi mã hóa khóa AES: %v", err)
	}

	// trả về kết quả
	return cipherTextBase64, encryptedAESKey, nil
}

// giải mã file
// ! AESKey truyền vào phải được giải mã trước đó
func RestoreFileFromNote(cipherTextBase64, decryptedAESKeyHex, outputFilePath string) error {

	// chuyển đổi AES Key từ base64 string -> byte[]
	aesKey, err := hex.DecodeString(decryptedAESKeyHex)
	if err != nil {
		return fmt.Errorf("lỗi định dạng khóa AES (không phải Hex hợp lệ): %v", err)
	}

	// chuyển chuỗi Base64 thành File Mã Hóa
	tempEncryptedPath := outputFilePath + ".enc_temp"
	defer os.Remove(tempEncryptedPath)

	// base64 String -> byte[]
	err = ConvertStringToBinary(cipherTextBase64, tempEncryptedPath)
	if err != nil {
		return fmt.Errorf("lỗi chuyển đổi Base64 sang file tạm: %v", err)
	}

	// giải mã
	err = DecryptedByAESKey(aesKey, tempEncryptedPath, outputFilePath)
	if err != nil {
		return fmt.Errorf("lỗi giải mã file (kiểm tra lại Key hoặc độ toàn vẹn file): %v", err)
	}

	fmt.Println("giải mã thành công")
	return nil
}

// Hỗ trợ test AES trên RAM

// Xử lý mã hóa và giải mã AES-GCM tối ưu trên RAM
// EncryptBytes mã hóa mảng byte bằng AES-GCM
// Output: [Nonce (12 bytes)] + [Ciphertext]
func EncryptBytes(plaintext []byte, key []byte) ([]byte, error) {
	// Tạo Block Cipher từ Key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Tạo GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Tạo Nonce ngẫu nhiên (Standard 12 bytes)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Mã hóa (Seal)
	// Kết quả = [Nonce] + [Ciphertext]
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// DecryptBytes giải mã mảng byte bằng AES-GCM
// Input: [Nonce (12 bytes)] + [Ciphertext]
func DecryptBytes(ciphertext []byte, key []byte) ([]byte, error) {
	// Tạo Block Cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Tạo GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Tách Nonce và Ciphertext thật
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("dữ liệu quá ngắn, không đúng định dạng AES-GCM")
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Giải mã (Open)
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, errors.New("giải mã thất bại (sai khóa hoặc dữ liệu bị sửa đổi)")
	}

	return plaintext, nil
}
