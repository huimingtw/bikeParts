package service

import (
	"strconv"

	"github.com/huimingtw/bikeparts/config"
	gomail "gopkg.in/gomail.v2"
)

type EmailService interface {
	Send(subject, body string) error
}

type EmailServiceImpl struct {
	cfg *config.AppConfig
}

func NewEmailService(cfg *config.AppConfig) *EmailServiceImpl {
	return &EmailServiceImpl{cfg: cfg}
}

func (e *EmailServiceImpl) Send(subject, body string) error {
	s := e.cfg.Get()

	port, err := strconv.Atoi(s.SMTPPort)
	if err != nil {
		port = 587
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.EmailUser)
	message.SetHeader("To", s.EmailTo)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	dialer := gomail.NewDialer("smtp.gmail.com", port, s.EmailUser, s.EmailPass)
	return dialer.DialAndSend(message)
}
