package handler

import (
	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/service"
)

type Handler struct {
	db     *db.DB
	mailer service.EmailService
}

func NewHandler(
	db *db.DB,
	mailer service.EmailService,
) *Handler {
	return &Handler{
		db,
		mailer,
	}
}
