package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

func SendSignupEmail(to, userName string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	email := Email{
		From:    os.Getenv("GMAIL_EMAIL"),
		To:      to,
		Subject: "Welcome to DesignMyPDF!",
		Body:    fmt.Sprintf(" Dear %s Thank you for signing up for DesignMyPDF! ", userName),
	}

	return sendEmail(email)
}

func SendForgotPasswordEmail(to, resetPasswordLink string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	email := Email{
		From:    os.Getenv("GMAIL_EMAIL"),
		To:      to,
		Subject: "DesignMyPDF Password Reset",
		Body:    fmt.Sprintf("To reset your password, click the following link: %s", resetPasswordLink),
	}

	return sendEmail(email)
}

func SendOTPEmail(to, otp string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	email := Email{
		From:    os.Getenv("GMAIL_EMAIL"),
		To:      to,
		Subject: "DesignMyPDF One-Time Password",
		Body:    fmt.Sprintf("Your one-time password for DesignMyPDF is: %s", otp),
	}

	return sendEmail(email)
}

func sendEmail(email Email) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	auth := smtp.PlainAuth("", os.Getenv("GMAIL_EMAIL"), os.Getenv("GMAIL_PASSWORD"), "smtp.gmail.com")

	msg := []byte("To: " + email.To + "\r\n" +
		"Subject: " + email.Subject + "\r\n" +
		"\r\n" +
		email.Body + "\r\n")

	err = smtp.SendMail("smtp.gmail.com:587", auth, email.From, []string{email.To}, msg)
	if err != nil {
		return err
	}
	return nil
}
