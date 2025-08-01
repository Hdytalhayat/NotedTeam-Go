// utils/mailer.go
package utils

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendVerificationEmail(toEmail, token string) error {
	host := os.Getenv("SMTP_HOST")
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	sender := os.Getenv("SMTP_SENDER_EMAIL")
	password := os.Getenv("SMTP_SENDER_PASSWORD")

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", sender)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", "Verify Your NotedTeam Account")

	// Buat link verifikasi
	verificationLink := "http://noble-energy-production-d0ae.up.railway.app:8080/auth/verify?token=" + token

	// Body email (bisa menggunakan HTML)
	body := "Hi there,<br><br>Thank you for registering. Please click the link below to verify your email address:<br>"
	body += "<a href=\"" + verificationLink + "\">Verify My Email</a><br><br>"
	body += "If you did not register for this account, you can safely ignore this email."

	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(host, port, sender, password)

	log.Printf("Sending verification email to %s", toEmail)
	if err := dialer.DialAndSend(mailer); err != nil {
		log.Printf("Failed to send email: %s", err)
		return err
	}

	return nil
}

func SendPasswordResetEmail(toEmail, token string) error {
	host := os.Getenv("SMTP_HOST")
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	sender := os.Getenv("SMTP_SENDER_EMAIL")
	password := os.Getenv("SMTP_SENDER_PASSWORD")

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", sender)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", "Reset Your NotedTeam Password")

	// Buat link ke halaman web reset
	resetLink := "http://192.168.1.2:8080/auth/reset-password-page?token=" + token

	body := "Hi there,<br><br>We received a request to reset your password. Please click the link below to set a new password:<br>"
	body += "<a href=\"" + resetLink + "\">Reset My Password</a><br><br>"
	body += "This link will expire in 1 hour. If you did not request a password reset, please ignore this email."

	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(host, port, sender, password)
	log.Printf("Sending verification email to %s", toEmail)
	if err := dialer.DialAndSend(mailer); err != nil {
		log.Printf("Failed to send email: %s", err)
		return err
	}
	return nil
}
