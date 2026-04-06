package handler

import (
	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/service"
)

type Handler struct {
	db       *db.DB
	notifier *service.NotificationService
}

func NewHandler(
	db *db.DB,
	notifier *service.NotificationService,
) *Handler {
	return &Handler{
		db:       db,
		notifier: notifier,
	}
}
