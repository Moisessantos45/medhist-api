package pkg

import (
	"api_citas/config"
	"context"
	"fmt"
	"log"
	"net/smtp"
)

// SendEmail es una función genérica para enviar cualquier correo HTML
func SendEmail(ctx context.Context, to []string, subject string, htmlBody string) error {
	emailConfig := config.GetEmailConfig()

	msg := fmt.Appendf(nil, "To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
		"%s",
		to[0], emailConfig.From, subject, htmlBody)

	auth := smtp.PlainAuth("", emailConfig.SMTPUser, emailConfig.SMTPPass, emailConfig.SMTPHost)

	err := smtp.SendMail(emailConfig.SMTPHost+":"+emailConfig.SMTPPort, auth, emailConfig.From, to, msg)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	return nil
}
