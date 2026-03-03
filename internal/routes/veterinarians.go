package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/features/veterinarians"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func VeterinariansRoutes(rg *gin.RouterGroup) {
	pR := veterinarians.NewPostgresRepository(db.DB)
	rd := config.Rdb
	maker := pkg.NewPasetoMaker()
	pS := veterinarians.NewVeterinarianUseCase(pR, rd, maker)
	pH := veterinarians.NewVeterinarianHandler(pS)

	rg.POST("/veterinarian/register", pH.Create)

	protected := rg.Group("/veterinarian")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("", pH.GetAll)
		protected.GET("/:id", pH.GetByID)
		protected.GET("/session", pH.GetByIdSession)
		protected.POST("", pH.Create)
		protected.PUT("/:id", pH.Update)
		protected.DELETE("/:id", pH.Delete)
	}

}
