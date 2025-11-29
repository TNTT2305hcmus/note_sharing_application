package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"note_sharing_application/server/services"
	"strings"
)

/*
	Sau khi người dùng đăng kí và đăng nhập thành công, server sẽ cấp cho một access token có thời hạn, hết hạn sẽ không cho
	phép tiếp tục thực hiện các chức năng, phải đăng nhập lại để nhận token mới

	Khi người dùng thực hiện gọi API sẽ gửi token (header) kèm theo thông tin (body):
	Chuẩn:
	Authorization: Bearer <token>
	Content-Type: application/json
	{
    "title": "Ghi chú",
    "content": "Nội dung đã mã hóa..."
	}

*/

// AuthMiddleware là hàm trả về một Gin Handler
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Kiểm tra định dạng: Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng Token sai (phải là Bearer <token>)"})
			c.Abort()
			return
		}

		//Lấy token
		token := parts[1]

		claims, err := services.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		//Lưu thông tin xác thực thu được từ token vào Context để sử dụng
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
