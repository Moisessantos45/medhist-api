package patients

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	s models.PatientUseCase
}

func NewPatientHandler(s models.PatientUseCase) *PatientHandler {
	return &PatientHandler{
		s: s,
	}
}

func (h *PatientHandler) GetAll(c *gin.Context) {
	page, pageSize, err := pkg.ValidateQueryPagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	results, err := h.s.GetAll(c, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, results)
}

func (h *PatientHandler) GetAllByVeterinarianID(c *gin.Context) {
	veterinarianID := c.MustGet("userID").(uint64)
	page, pageSize, err := pkg.ValidateQueryPagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Println("Veterinarian page:", page, "pageSize:", pageSize)

	results, err := h.s.GetAllByVeterinarianID(c, veterinarianID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, results)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, err := h.s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": result})
}

func (h *PatientHandler) Create(c *gin.Context) {
	id := c.MustGet("userID").(uint64)
	var patient models.Patient

	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if id != patient.VeterinarianID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Veterinarian ID mismatch"})
		return
	}

	log.Println(id, patient)

	patient.VeterinarianID = id

	err := h.s.Create(c, &patient)
	if err != nil {
		if strings.Contains(err.Error(), "ya se encuentra registrado") {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Patient created successfully", "data": patient})
}

func (h *PatientHandler) Update(c *gin.Context) {

	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var patient models.Patient

	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Update(c, id, &patient)
	if err != nil {
		if strings.Contains(err.Error(), "ya se encuentra registrado") {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Patient updated successfully", "data": patient})
}

func (h *PatientHandler) UpdateStatus(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.UpdateStatus(id, body.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Patient status updated successfully"})
}

func (h *PatientHandler) Delete(c *gin.Context) {
	vetId := c.MustGet("userID").(uint64)

	patientId, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Delete(patientId, vetId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Patient deleted successfully"})
}
