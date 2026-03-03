package medicalrecords

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MedicalRecordHandler struct {
	s models.MedicalRecordUseCase
}

func NewMedicalRecordHandler(s models.MedicalRecordUseCase) *MedicalRecordHandler {
	return &MedicalRecordHandler{s: s}
}

func (h *MedicalRecordHandler) GetAll(c *gin.Context) {
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

func (h *MedicalRecordHandler) GetByID(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *MedicalRecordHandler) Create(c *gin.Context) {
	id := c.MustGet("userID").(uint64)
	var req models.MedicalRecord

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ BIND ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if id != req.VeterinarianID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not authorized to create a medical record for this veterinarian"})
		return
	}

	err := h.s.Create(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Medical record created successfully",
		"data":    req,
	})
}

func (h *MedicalRecordHandler) Update(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var req models.MedicalRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Update(c, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Medical record updated successfully",
		"data":    req,
	})
}

func (h *MedicalRecordHandler) Delete(c *gin.Context) {
	vetID := c.MustGet("userID").(uint64)

	medicalRecordID, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Delete(medicalRecordID, vetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Medical record deleted successfully",
	})
}
