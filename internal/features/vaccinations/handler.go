package vaccinations

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type VaccinationHandler struct {
	s models.VaccinationUseCase
}

func NewVaccinationHandler(s models.VaccinationUseCase) *VaccinationHandler {
	return &VaccinationHandler{s: s}
}

func (h *VaccinationHandler) GetAll(c *gin.Context) {
	vertinarianId, err := pkg.ValidateParamsId(c, "veterinarian_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	patientId, err := pkg.ValidateParamsId(c, "patient_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	page, pageSize, err := pkg.ValidateQueryPagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	results, err := h.s.GetAll(c, patientId, vertinarianId, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, results)
}

func (h *VaccinationHandler) GetByID(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	vaccination, err := h.s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": vaccination})
}

func (h *VaccinationHandler) Create(c *gin.Context) {
	id := c.MustGet("userID").(uint64)
	var vaccination models.Vaccination

	if err := c.ShouldBindJSON(&vaccination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if id != vaccination.VeterinarianID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not authorized to create a vaccination for this veterinarian"})
		return
	}

	err := h.s.Create(c, &vaccination)
	if err != nil {
		log.Printf("Error creating vaccination: %v", err)
		if strings.Contains(err.Error(), "ya se encuentra registrada") {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Vaccination created successfully", "data": vaccination})
}

func (h *VaccinationHandler) Update(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var vaccination models.Vaccination

	if err := c.ShouldBindJSON(&vaccination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Update(c, id, &vaccination)
	if err != nil {
		if strings.Contains(err.Error(), "ya se encuentra registrada") {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Vaccination updated successfully", "data": vaccination})
}

func (h *VaccinationHandler) Delete(c *gin.Context) {
	vetId := c.MustGet("userID").(uint64)

	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Delete(id, vetId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Vaccination deleted successfully"})
}
