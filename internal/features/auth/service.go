package auth

import (
	"api_citas/internal/pkg"
	"api_citas/internal/pkg/errorsx"
	"api_citas/internal/pkg/models"
	"api_citas/internal/templates"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthUseCase struct {
	v      models.VeterinarianUseCase
	rd     *redis.Client
	marker *pkg.PasetoMaker
}

func NewAuthUseCase(v models.VeterinarianUseCase, rd *redis.Client, mk *pkg.PasetoMaker) *AuthUseCase {
	return &AuthUseCase{v: v, rd: rd, marker: mk}
}

func (a *AuthUseCase) Login(ctx context.Context, email string, password string) (*models.Veterinarian, error) {
	veterinarian, err := a.v.GetActiveUserByEmail(email)
	if err != nil {
		if errors.Is(err, errorsx.ErrCodeEmailNotVerified) {
			log.Printf("Intento de inicio de sesión con email no verificado: %s", email)
			return nil, errorsx.ErrCodeEmailNotVerified
		}

		log.Printf("Error fetching user by email: %v", err)
		return nil, fmt.Errorf("No se encontro email")
	}

	if !pkg.CheckPasswordHash(password, veterinarian.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	token, err := a.marker.NewToken(fmt.Sprintf("%d", veterinarian.ID), 8*time.Hour)
	if err != nil {
		return nil, err
	}

	err = a.rd.Set(ctx, token, veterinarian.ID, 8*time.Hour).Err()

	if err != nil {
		return nil, err
	}

	veterinarian.Token = token
	veterinarian.Password = ""

	return veterinarian, nil
}

func (a *AuthUseCase) GetSession(id uint64) (*models.Veterinarian, error) {
	result, err := a.v.GetByID(id)
	if err != nil {
		return nil, err
	}

	result.Password = ""

	return result, err
}

func (a *AuthUseCase) ConfirmAccount(ctx context.Context, id uint64, token string) error {

	log.Printf("Confirming account for user ID %d with token: %s", id, token[:8]+"...")
	deleted, err := a.rd.Del(ctx, token).Result()
	if err != nil {
		return fmt.Errorf("error de conexión con sesión: %w", err)
	}

	if deleted == 0 {
		return fmt.Errorf("el enlace ya fue utilizado o ha expirado")
	}

	err = a.v.UpdateEmailConfirmed(ctx, id)
	if err != nil {
		return fmt.Errorf("error al activar la cuenta en base de datos: %w", err)
	}

	return nil
}

func (a *AuthUseCase) SendPasswordReset(ctx context.Context, email string) error {
	isProduction := os.Getenv("GO_ENV")
	var host = os.Getenv("HOST_URL_PROD")
	if isProduction == "dev" {
		host = os.Getenv("HOST_URL_DEV")
	}

	veterinarian, err := a.v.GetByEmail(email)
	if err != nil {
		return err
	}

	token, err := a.marker.NewToken(fmt.Sprintf("%d", veterinarian.ID), 15*time.Minute)
	if err != nil {
		return err
	}

	err = a.v.UpdateToken(veterinarian.ID, token)
	if err != nil {
		return fmt.Errorf("error updating token in database: %w", err)
	}

	err = a.rd.Set(ctx, token, fmt.Sprintf("%s_%d_%s", "ps", veterinarian.ID, token), 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error caching token: %w", err)
	}

	renderer, err := templates.NewEmailRenderer()
	if err != nil {
		return err
	}

	data := templates.PasswordResetData{
		Name:      veterinarian.Name,
		ResetLink: fmt.Sprintf("%s/forgot-password/%s", host, token),
	}

	htmlContent, err := renderer.RenderPasswordReset(data)
	if err != nil {
		return err
	}

	err = pkg.SendEmail(ctx, []string{email}, "Restablece tu contraseña", htmlContent)

	return err
}

func (a *AuthUseCase) ForwardEmailVerification(ctx context.Context, email string) error {
	if email == "" || len(email) <= 10 || !pkg.EmailRegex.MatchString(email) {
		return fmt.Errorf("email inválido: %s", email)
	}

	user, err := a.v.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("No se encontro email")
	}

	if user.EmailConfirmed {
		return fmt.Errorf("El correo electrónico ya ha sido confirmado")
	}

	isProduction := os.Getenv("GO_ENV")
	var host = os.Getenv("HOST_URL_PROD")
	if isProduction == "dev" {
		host = os.Getenv("HOST_URL_DEV")
	}

	token, err := a.marker.NewToken(fmt.Sprintf("%d", user.ID), 15*time.Minute)
	if err != nil {
		return err
	}

	err = a.rd.Set(ctx, token, user.ID, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error caching token: %w", err)
	}

	renderer, err := templates.NewEmailRenderer()
	if err != nil {
		return err
	}

	data := templates.ConfirmAccountData{
		Name:        user.Name,
		ConfirmLink: fmt.Sprintf("%s/confirm/%s", host, token),
	}

	htmlContent, err := renderer.RenderConfirmAccount(data)
	if err != nil {
		return err
	}

	err = pkg.SendEmail(ctx, []string{email}, "Confirmacion de cuenta", htmlContent)
	if err != nil {
		return fmt.Errorf("No se pudo enviar el correo electrónico de confirmación: %w", err)
	}

	return nil
}

func (a *AuthUseCase) ChangePassword(ctx context.Context, id uint64, currentPassword string, newPassword string) error {
	return a.v.ChangePassword(id, currentPassword, newPassword)
}

func (a *AuthUseCase) ResetPassword(ctx context.Context, id uint64, token string, newPassword string) error {
	err := a.v.ResetPassword(ctx, id, token, newPassword)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = a.rd.Del(ctx, fmt.Sprintf("%s_%d_%s", "ps", id, token)).Err()
	if err != nil {
		log.Printf("Error deleting password reset token for user %d: %v", id, err)
	}

	return nil
}

func (a *AuthUseCase) Logout(ctx context.Context, token string) error {
	log.Printf("Attempting to logout with token: %s", token[:8]+"...")
	if token == "" || len(token) < 7 {
		return fmt.Errorf("missing or invalid token")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	deleted, err := a.rd.Del(ctx, token).Result()
	if err != nil {
		log.Printf("Redis DEL error for token %s: %v", token[:8]+"...", err)
		return fmt.Errorf("failed to delete token: %w", err)
	}

	if deleted == 0 {
		log.Printf("Token not found for deletion: %s", token[:8]+"...")
		return nil
	}

	return nil
}
