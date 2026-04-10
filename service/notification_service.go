package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/huimingtw/bikeparts/models"
)

type dbQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type NotificationService struct {
	mailer EmailService
	logger *slog.Logger
}

func NewNotificationService(mailer EmailService, logger *slog.Logger) *NotificationService {
	return &NotificationService{mailer: mailer, logger: logger}
}

func (n *NotificationService) CheckAndNotify(ctx context.Context, db dbQuerier, part models.Part) error {
	if part.Stock > part.ReorderLevel {
		n.logger.DebugContext(ctx, "stock above reorder level, skip notification", "part_id", part.ID, "sku", part.SKU, "stock", part.Stock, "reorder_level", part.ReorderLevel)
		return nil
	}

	exists := false
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM low_stock_notifications WHERE part_id = ? AND deleted_at IS NULL)",
		part.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		n.logger.DebugContext(ctx, "notification already exists, skip", "part_id", part.ID, "sku", part.SKU)
		return nil
	}

	_, err = db.ExecContext(ctx,
		"INSERT INTO low_stock_notifications (part_id, created_at) VALUES (?, ?)",
		part.ID, time.Now())
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("[Low Stock] %s (SKU: %s) only %d left", part.Name, part.SKU, part.Stock)
	body := fmt.Sprintf("Part %s (SKU: %s) is low on stock: %d remaining (reorder level: %d).", part.Name, part.SKU, part.Stock, part.ReorderLevel)
	if err := n.mailer.Send(subject, body); err != nil {
		n.logger.ErrorContext(ctx, "failed to send low stock email", "part_id", part.ID, "sku", part.SKU, "err", err)
		return err
	}
	n.logger.InfoContext(ctx, "low stock email sent", "part_id", part.ID, "sku", part.SKU, "stock", part.Stock)
	return nil
}

func (n *NotificationService) ClearNotification(ctx context.Context, db dbQuerier, partID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE low_stock_notifications SET deleted_at = ? WHERE part_id = ? AND deleted_at IS NULL",
		time.Now(), partID)
	return err
}
