package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/huimingtw/bikeparts/models"
)

type NotificationService struct {
	mailer EmailService
	logger *slog.Logger
}

func NewNotificationService(mailer EmailService, logger *slog.Logger) *NotificationService {
	return &NotificationService{mailer: mailer, logger: logger}
}

func (n *NotificationService) CheckAndNotify(ctx context.Context, db *sql.DB, part models.Part) error {
	if part.Stock > part.ReorderLevel {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	exists := false
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM low_stock_notifications WHERE part_id = ? AND deleted_at IS NULL)", part.ID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = tx.Exec("INSERT INTO low_stock_notifications (part_id, created_at) VALUES (?, ?)", part.ID, time.Now())
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("[Low Stock] %s (SKU: %s) only %d left", part.Name, part.SKU, part.Stock)
	body := fmt.Sprintf("Part %s (SKU: %s) is low on stock: %d remaining (reorder level: %d).", part.Name, part.SKU, part.Stock, part.ReorderLevel)
	if err := n.mailer.Send(subject, body); err != nil {
		n.logger.ErrorContext(ctx, "failed to send low stock email", "part_id", part.ID, "sku", part.SKU, "err", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	n.logger.InfoContext(ctx, "low stock email sent", "part_id", part.ID, "sku", part.SKU, "stock", part.Stock)
	return nil
}

func (n *NotificationService) ClearNotification(db *sql.DB, partID int64) error {
	_, err := db.Exec("UPDATE low_stock_notifications SET deleted_at = ? WHERE part_id = ? AND deleted_at IS NULL", time.Now(), partID)
	return err
}
