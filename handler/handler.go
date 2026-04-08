package handler

import (
	"database/sql"
	"log/slog"

	"github.com/huimingtw/bikeparts/service"
)

type Handler struct {
	db       *sql.DB
	notifier *service.NotificationService
	logger   *slog.Logger
}

func NewHandler(
	db *sql.DB,
	notifier *service.NotificationService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		db:       db,
		notifier: notifier,
		logger:   logger,
	}
}
