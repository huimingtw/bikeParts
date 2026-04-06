package service

import (
	"os"
	"strconv"

	gomail "gopkg.in/gomail.v2"
)

type EmailService interface {
	Send(subject, body string) error
}

type EmailServiceImpl struct {
	user    string
	pass    string
	mail_to string
	port    int
}

func NewEmailService() *EmailServiceImpl {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		port = 587
	}
	return &EmailServiceImpl{
		user:    os.Getenv("EMAIL_USER"),
		pass:    os.Getenv("EMAIL_PASS"),
		mail_to: os.Getenv("EMAIL_TO"),
		port:    port,
	}
}

func (e *EmailServiceImpl) Send(subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", e.user)
	message.SetHeader("To", e.mail_to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	dialer := gomail.NewDialer("smtp.gmail.com", e.port, e.user, e.pass)
	return dialer.DialAndSend(message)
}
