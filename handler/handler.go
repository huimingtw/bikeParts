package handler

import (
	"log/slog"

	"github.com/huimingtw/bikeparts/db"
	"github.com/huimingtw/bikeparts/service"
)

type Handler struct {
	db       *db.DB
	notifier *service.NotificationService
	logger   *slog.Logger
}

func NewHandler(
	db *db.DB,
	notifier *service.NotificationService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		db:       db,
		notifier: notifier,
		logger:   logger,
	}
}
