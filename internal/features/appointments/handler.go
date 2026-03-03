package appointments

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	s models.AppointmentUseCase
}

func NewAppointmentHandler(s models.AppointmentUseCase) *AppointmentHandler {
	return &AppointmentHandler{
		s: s,
	}
}

func (h *AppointmentHandler) GetAll(c *gin.Context) {
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

func (h *AppointmentHandler) GetByID(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": result})
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	id := c.MustGet("userID").(uint64)

	var input models.Appointment

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if id != input.VeterinarianID {
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not authorized to create an appointment for this veterinarian"})
		return
	}

	err := h.s.Create(c, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Appointment created successfully", "data": input})
}

func (h *AppointmentHandler) Update(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var input models.Appointment

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.s.Update(c, id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Appointment updated successfully", "data": input})
}

func (h *AppointmentHandler) Delete(c *gin.Context) {
	vetID := c.MustGet("userID").(uint64)

	appointmentID, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err = h.s.Delete(appointmentID, vetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": "Appointment deleted successfully"})
}
