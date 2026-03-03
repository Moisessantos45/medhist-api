package veterinarians

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VeterinarianHandler struct {
	s models.VeterinarianUseCase
}

func NewVeterinarianHandler(s models.VeterinarianUseCase) *VeterinarianHandler {
	return &VeterinarianHandler{
		s: s,
	}
}

func (h *VeterinarianHandler) GetAll(c *gin.Context) {
	page, pageSize, err := pkg.ValidateQueryPagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	results, err := h.s.GetAll(c, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *VeterinarianHandler) GetByIdSession(c *gin.Context) {
	id := c.MustGet("userID").(uint64)

	result, err := h.s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *VeterinarianHandler) GetByID(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	result, err := h.s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *VeterinarianHandler) GetByEmail(c *gin.Context) {

	email := c.Query("email")
	result, err := h.s.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *VeterinarianHandler) Create(c *gin.Context) {

	var veterinarian models.Veterinarian
	if err := c.ShouldBindJSON(&veterinarian); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err := h.s.Create(c, &veterinarian)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": veterinarian,
	})
}

func (h *VeterinarianHandler) Update(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var veterinarian models.Veterinarian
	if err := c.ShouldBindJSON(&veterinarian); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = h.s.Update(id, &veterinarian)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Veterinarian updated successfully",
		"data":    veterinarian,
	})
}

func (h *VeterinarianHandler) Delete(c *gin.Context) {
	id, err := pkg.ValidateParamsId(c, "")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = h.s.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Veterinarian deleted successfully",
	})

}
