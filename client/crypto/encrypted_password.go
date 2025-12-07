package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// Hàm mã hóa Password bằng Server Public Key RSA
func EncryptPasswordWithServerKey(password string, pubKeyPEM string) (string, error) {
	// Decode khối PEM
	block, _ := pem.Decode([]byte(pubKeyPEM))
	if block == nil {
		return "", errors.New("Can't get server public key RSA")
	}

	// Parse thành RSA Public Key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("Not RSA key")
	}

	// Mã hóa Password bằng EncryptedOAEP
	// Server giải mã bằng DecryptedOAEP
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaPub,
		[]byte(password),
		nil,
	)
	if err != nil {
		return "", err
	}

	// Trả về Base64 để gửi qua JSON
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
