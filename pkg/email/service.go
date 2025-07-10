package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type Email struct {
	From        string
	To          string
	Subject     string
	Body        string
	ContentType string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
}

func SendSignupEmail(to, userName string) error {
	email := Email{
		From:    os.Getenv("GMAIL_EMAIL"),
		To:      to,
		Subject: "Welcome to DesignMyPDF!",
		Body:    fmt.Sprintf("Dear %s, Thank you for signing up for DesignMyPDF!", userName),
	}

	return sendEmail(email)
}

func SendForgotPasswordEmail(to, token string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	link := "http://localhost:3000/reset-password"
	stage := os.Getenv("GO_ENV")
	if stage == "production" {
		link = "https://designmypdf.vercel.app/reset-password"
	}
	email := Email{
		From:        os.Getenv("GMAIL_EMAIL"),
		To:          to,
		Subject:     "DesignMyPDF Password Reset",
		Body:        fmt.Sprintf("<p>To reset your password, click the following link:</p><p><a href='%s?token=%s'>Reset Password</a></p>", link, token),
		ContentType: "text/html",
	}

	return sendEmail(email)
}

func sendEmail(email Email) error {
	auth := smtp.PlainAuth("", os.Getenv("GMAIL_EMAIL"), os.Getenv("GMAIL_PASSWORD"), "smtp.gmail.com")

	headers := "To: " + email.To + "\r\n" +
		"Subject: " + email.Subject + "\r\n"

	if email.ContentType == "text/html" {
		headers += "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	}

	msg := []byte(headers + "\r\n" + email.Body + "\r\n")

	err := smtp.SendMail("smtp.gmail.com:587", auth, email.From, []string{email.To}, msg)
	if err != nil {
		return err
	}
	return nil
}
