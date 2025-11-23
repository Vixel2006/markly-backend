package services

import (
	"os"

	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendEmail(to, subject, msg string) error
}

type emailService struct {
	from string
}

func NewEmailService() EmailService {
	return &emailService{
		from: os.Getenv("SMTP_USERNAME"),
	}
}

func (e *emailService) SendEmail(to, subject, msg string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", e.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", msg)

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
