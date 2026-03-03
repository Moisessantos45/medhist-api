package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/features/patients"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func PatientsRoutes(rg *gin.RouterGroup) {
	pR := patients.NewPostgresRepository(db.DB)
	rd := config.Rdb
	pS := patients.NewPatientUseCase(pR, rd)
	pH := patients.NewPatientHandler(pS)

	maker := pkg.NewPasetoMaker()

	protected := rg.Group("/patient")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("", pH.GetAll)
		protected.GET("/veterinarian", pH.GetAllByVeterinarianID)
		protected.GET("/:id", pH.GetByID)
		protected.POST("", pH.Create)
		protected.PUT("/:id", pH.Update)
		protected.PATCH("/:id/status", pH.UpdateStatus)
		// protected.DELETE("/:id", pH.Delete)
	}

}
