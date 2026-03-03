package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	a AuthUseCase
}

func NewAuthHandler(a AuthUseCase) *AuthHandler {
	return &AuthHandler{a: a}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	veterinarian, err := h.a.Login(c, input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": veterinarian})
}

func (h *AuthHandler) GetSession(c *gin.Context) {
	id := c.MustGet("userID").(uint64)

	veterinarian, err := h.a.GetSession(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": veterinarian})
}

func (h *AuthHandler) ConfirmAccount(c *gin.Context) {
	id := c.MustGet("userID").(uint64)
	token := c.MustGet("token").(string)

	err := h.a.ConfirmAccount(c, id, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account confirmed successfully"})
}

func (h *AuthHandler) SendResetPasswordEmail(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := h.a.SendPasswordReset(c, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reset password email sent successfully"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	id := c.MustGet("userID").(uint64)

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=7"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := h.a.ChangePassword(c, id, input.CurrentPassword, input.NewPassword)
	if err != nil {
		log.Printf("Error changing password for user %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	id := c.MustGet("userID").(uint64)
	token := c.MustGet("token").(string)

	var input struct {
		NewPassword string `json:"new_password" binding:"required,min=7"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := h.a.ResetPassword(c, id, token, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.MustGet("token").(string)

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing token"})
		return
	}

	err := h.a.Logout(c, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
