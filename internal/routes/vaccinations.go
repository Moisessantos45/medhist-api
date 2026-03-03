package routes

import (
	"api_citas/config"
	"api_citas/config/db"
	"api_citas/internal/features/vaccinations"
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func VaccinationsRoutes(rg *gin.RouterGroup) {
	vR := vaccinations.NewPostgresRepository(db.DB)
	rd := config.Rdb
	vS := vaccinations.NewVaccinationUseCase(vR, rd)
	vH := vaccinations.NewVaccinationHandler(vS)

	maker := pkg.NewPasetoMaker()

	protected := rg.Group("/vaccination")
	protected.Use(middleware.AuthMiddleware(maker, rd))
	{
		protected.GET("/patient/:patient_id/veterinarian/:veterinarian_id", vH.GetAll)
		protected.GET("/:id", vH.GetByID)
		protected.POST("", vH.Create)
		protected.PUT("/:id", vH.Update)
		protected.DELETE("/:id", vH.Delete)
	}

}
