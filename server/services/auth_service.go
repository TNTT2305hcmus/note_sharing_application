package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
	Access Token dùng để xác thực người dùng với thông tin (user_id và username) được gói trong Claims
	Khi người dùng xác thực thành công thông tin đăng nhập sẽ gửi kèm Access Token về cho người dùng
	Mỗi khi người dùng gọi API cần phải có token xác thực để server biết có đúng người dùng hay không
	Nếu token hết hạn, phải đăng nhập lại để nhận access token mới

	**Chú ý bên client cần xử lý khi access token hết hạn tránh gây khó khăn khi đang sử dụng
	mà hết hạn access token
*/

// Định nghĩa một JWTClaims với thông tin user và thông tin chuẩn của Claims
type authClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Chỉ lấy secretKey từ file .env khi cần
func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	return []byte(secret)
}

// Tạo access token xác thực khi đăng nhập thành công
func GenerateAuthJWT(userID, username string) (string, error) {
	//Quy định thời gian hết hạn của token
	expirationTime := time.Now().Add(15 * time.Minute)

	//Ghi nội dung cho claims
	claims := &authClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	//Tạo token với nội dung đã ghi với thuật toán mã hóa HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Kí với secretKey của server và trả về token đã được kí
	return token.SignedString(getSecretKey())
}

// Kiểm tra token
func ValidateAuthJWT(tokenString string) (*authClaims, error) {
	//Dịch token từ chuỗi đầu vào
	token, err := jwt.ParseWithClaims(tokenString, &authClaims{},
		//Hàm callback kiểm tra thuật toán mã hóa có đúng với thuật toán server đã dùng hay không
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("thuật toán không xác định: %v", token.Header["alg"])
			}
			//Nếu đúng thì lấy secretKey để giải mã token
			return getSecretKey(), nil
		})

	//Nếu có lỗi thì báo lỗi
	if err != nil {
		return nil, err
	}

	/*
		Nếu token đã hợp lệ (đúng signature và còn hạn) thì lấy thông tin
		từ token bằng cách ép kiểu sang JWTClaims đã định nghĩa để
		có thể sử dụng dễ dàng
	*/
	if claims, ok := token.Claims.(*authClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
