package handler

import (
	"database/sql"
	"log/slog"

	"github.com/huimingtw/bikeparts/config"
	"github.com/huimingtw/bikeparts/service"
)

type Handler struct {
	db       *sql.DB
	notifier *service.NotificationService
	logger   *slog.Logger
	cfg      *config.AppConfig
	mailer   service.EmailService
}

func NewHandler(
	db *sql.DB,
	notifier *service.NotificationService,
	logger *slog.Logger,
	cfg *config.AppConfig,
	mailer service.EmailService,
) *Handler {
	return &Handler{
		db:       db,
		notifier: notifier,
		logger:   logger,
		cfg:      cfg,
		mailer:   mailer,
	}
}
