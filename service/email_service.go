package service

import (
	"fmt"
	"os"
	"strconv"

	"github.com/huimingtw/bikeparts/models"
	gomail "gopkg.in/gomail.v2"
)

type EmailService interface {
	SendLowStockEmail(part models.Part) error
}

type EmailServiceImpl struct{}

func (e *EmailServiceImpl) SendLowStockEmail(part models.Part) error {
	user := os.Getenv("EMAIL_USER")
	pass := os.Getenv("EMAIL_PASS")
	to := os.Getenv("EMAIL_TO")
	portStr := os.Getenv("SMTP_PORT")
	port, error := strconv.Atoi(portStr)
	if error != nil {
		port = 587
	}

	message := gomail.NewMessage()
	message.SetHeader("From", user)
	message.SetHeader("To", to)
	message.SetHeader("Subject", fmt.Sprintf("[Low Stock] %s (SKU: %s) only %d left", part.Name, part.SKU, part.Stock))
	message.SetBody("text/plain", fmt.Sprintf("Part %s (SKU: %s) is low on stock: %d remaining (reorder level: %d).", part.Name, part.SKU, part.Stock, part.ReorderLevel))

	dialer := gomail.NewDialer("smtp.gmail.com", port, user, pass)
	return dialer.DialAndSend(message)
}
