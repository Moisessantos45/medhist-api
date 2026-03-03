package config

import "os"

type EmailConfig struct {
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	From     string
}

func GetEmailConfig() EmailConfig {

	smtpHost := os.Getenv("SMTP_HOST")    // Host SMTP de Brevo
	smtpPort := os.Getenv("SMTP_PORT")    // Puerto SMTP de Brevo
	smtpUser := os.Getenv("SMTP_USER")    // Email de login
	smtpPass := os.Getenv("API_KEY_SMTP") // SMTP key de dashboard
	from := "shigatsutranslations@gmail.com"

	return EmailConfig{
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
		SMTPUser: smtpUser,
		SMTPPass: smtpPass,
		From:     from,
	}
}
