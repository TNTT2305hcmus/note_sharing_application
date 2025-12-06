package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// Tạo biến toàn bộ khởi tạo chung khi server được khởi tạo
var (
	ServerPrivateKey *rsa.PrivateKey
	ServerPublicKey  *rsa.PublicKey
)

// Tạo cặp khóa RSA (private key và public key) của Server
func GenerateServerRSAKeys() error {
	var err error
	ServerPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	// Struct của private key thường chứa sẵn public key
	// Lấy địa chỉ của public key từ private key
	ServerPublicKey = &ServerPrivateKey.PublicKey
	return nil
}

// Chuyển Public Key sang dạng string (PEM format) để dễ gửi
// Trả về public key dưới dạng string (PEM)
func ExportPublicKeyAsPEM() (string, error) {
	// Chuyển public key từ struct GO sang dạng bytes (nhị phân)
	pubASN1, err := x509.MarshalPKIXPublicKey(ServerPublicKey)
	if err != nil {
		return "", err
	}

	// Mã hóa bytes thành định dạng PEM để gửi
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return string(pubBytes), nil
}

// Phương pháp OAEP (Optimal Asymmetric Encryption Padding)
// Là phương pháp giải mã an toàn cho RSA
// Client mã hóa password bằng EncrypteOAEP
func DecryptOAEP(cipherText []byte) (string, error) {
	hash := crypto.SHA256

	decryptedBytes, err := rsa.DecryptOAEP(
		hash.New(),
		rand.Reader,
		ServerPrivateKey,
		cipherText,
		nil,
	)
	if err != nil {
		return "", errors.New("Fail to decrypted")
	}

	return string(decryptedBytes), nil
}
