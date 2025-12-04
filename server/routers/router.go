package routers

import (
	"github.com/gin-gonic/gin"

	// Import các package nội bộ
	"note_sharing_application/server/handlers"
	"note_sharing_application/server/middlewares"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Cấu hình CORS (Nên có nếu Frontend và Backend khác port)
	// r.Use(cors.Default())

	api := r.Group("/")
	{
		// --- AUTH ROUTES (Giữ nguyên của bạn) ---
		api.POST("/register", handlers.RegisterHandler)
		api.POST("/login", handlers.LoginHandler)

		// --- NOTE ROUTES ---
		// Nhóm các API cần đăng nhập
		noteRoutes := api.Group("/notes")

		// Sử dụng Middleware xác thực người dùng (JWT)
		noteRoutes.Use(middlewares.AuthMiddleware())
		{
			// CRUD Note cơ bản (Giữ nguyên của bạn)
			noteRoutes.POST("", handlers.CreateNoteHandler) // Tạo note gốc
			noteRoutes.GET("", handlers.GetNoteHandler)     // Lấy danh sách note

			// --- URL SHARING ROUTES (Mới thêm vào) ---

			// 1. Tạo URL chia sẻ cho một Note cụ thể
			// Có thêm middleware: ValidateCreateUrl (check chủ sở hữu, check metadata)
			noteRoutes.POST("/:note_id/url", middlewares.ValidateCreateUrl(), handlers.CreateNoteUrl)

			//Nếu muốn xem thì cần tìm 1 url sẵn trước thì mới được truy cập
			noteRoutes.GET("/:note_id/url", middlewares.ValidateUrlAccess(), handlers.GetNoteUrl)

			//API xóa ghi chú, hủy chia sẻ viết ở đây
		}
		api.GET("/note/:url_id", middlewares.AuthMiddleware(), middlewares.ValidateUrlAccess(), middlewares.ValidateUrl(), handlers.ViewNoteHandler)
	}
	return r
}
