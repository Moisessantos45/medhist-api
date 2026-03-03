package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/features/auth"
	"api_citas/internal/features/veterinarians"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(rg *gin.RouterGroup) {
	vrepo := veterinarians.NewPostgresRepository(db.DB)
	rd := config.Rdb
	mk := pkg.NewPasetoMaker()
	vusecase := veterinarians.NewVeterinarianUseCase(vrepo, rd, mk)
	ausecase := auth.NewAuthUseCase(vusecase, rd, mk)
	authHandler := auth.NewAuthHandler(*ausecase)
	maker := pkg.NewPasetoMaker()

	rg.POST("/login", authHandler.Login)
	rg.POST("/forgot-password", authHandler.SendResetPasswordEmail)

	protected := rg.Group("/")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("/confirm-account", authHandler.ConfirmAccount)
		protected.GET("/session", authHandler.GetSession)
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/reset-password", authHandler.ResetPassword)
		protected.POST("/change-password", authHandler.ChangePassword)
	}
}
