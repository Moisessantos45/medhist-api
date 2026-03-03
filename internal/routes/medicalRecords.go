package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	medicalrecords "api_citas/internal/features/medical_records"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func MedicalRecordsRoutes(rg *gin.RouterGroup) {
	pR := medicalrecords.NewPostgresRepository(db.DB)
	rd := config.Rdb
	pS := medicalrecords.NewMedicalRecordUseCase(pR, rd)
	pH := medicalrecords.NewMedicalRecordHandler(pS)

	maker := pkg.NewPasetoMaker()

	protected := rg.Group("/medical-records")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("", pH.GetAll)
		protected.GET("/patient/:patient_id/veterinarian/:veterinarian_id", pH.GetAll)
		protected.GET("/:id", pH.GetByID)
		protected.POST("", pH.Create)
		protected.PUT("/:id", pH.Update)
		protected.DELETE("/:id", pH.Delete)
	}

}
