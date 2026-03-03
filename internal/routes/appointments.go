package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/features/appointments"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func AppointmentsRoutes(rg *gin.RouterGroup) {
	aR := appointments.NewPostgresRepository(db.DB)
	rd := config.Rdb
	aS := appointments.NewAppointmentUseCase(aR, rd)
	aH := appointments.NewAppointmentHandler(aS)
	maker := pkg.NewPasetoMaker()

	protected := rg.Group("/appointment")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("/patient/:patient_id/veterinarian/:veterinarian_id", aH.GetAll)
		protected.GET("/:id", aH.GetByID)
		protected.POST("", aH.Create)
		protected.PUT("/:id", aH.Update)
		protected.DELETE("/:id", aH.Delete)
	}
}
