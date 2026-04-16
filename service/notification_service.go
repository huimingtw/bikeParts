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
	db     *sql.DB
}

func NewNotificationService(mailer EmailService, logger *slog.Logger, db *sql.DB) *NotificationService {
	return &NotificationService{mailer: mailer, logger: logger, db: db}
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

	result, err := db.ExecContext(ctx,
		"INSERT INTO low_stock_notifications (part_id, created_at) VALUES (?, ?)",
		part.ID, time.Now())
	if err != nil {
		return err
	}
	notifID, _ := result.LastInsertId()

	subject := fmt.Sprintf("【庫存警示】%s（SKU: %s）剩餘 %d 件", part.Name, part.SKU, part.Stock)
	body := fmt.Sprintf("零件名稱：%s\nSKU：%s\n目前庫存：%d\n再訂購水位：%d\n\n請盡快補貨。", part.Name, part.SKU, part.Stock, part.ReorderLevel)
	go func() {
		if err := n.mailer.Send(subject, body); err != nil {
			n.logger.ErrorContext(context.Background(), "failed to send low stock email", "part_id", part.ID, "sku", part.SKU, "err", err)
			return
		}
		_, _ = n.db.Exec("UPDATE low_stock_notifications SET sent_at = ? WHERE id = ?", time.Now(), notifID)
		n.logger.InfoContext(context.Background(), "low stock email sent", "part_id", part.ID, "sku", part.SKU, "stock", part.Stock)
	}()

	return nil
}

func (n *NotificationService) ClearNotification(ctx context.Context, db dbQuerier, partID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE low_stock_notifications SET deleted_at = ? WHERE part_id = ? AND deleted_at IS NULL",
		time.Now(), partID)
	return err
}

// RetryUnsent looks for notifications where the email was never sent and retries them.
// Intended to be called on a schedule (e.g. every 5 minutes).
func (n *NotificationService) RetryUnsent(ctx context.Context) {
	rows, err := n.db.QueryContext(ctx, `
		SELECT n.id, p.name, p.sku, p.stock, p.reorder_level
		FROM low_stock_notifications n
		JOIN parts p ON p.id = n.part_id
		WHERE n.sent_at IS NULL AND n.deleted_at IS NULL
	`)
	if err != nil {
		n.logger.ErrorContext(ctx, "failed to query unsent notifications", "err", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var part models.Part
		if err := rows.Scan(&id, &part.Name, &part.SKU, &part.Stock, &part.ReorderLevel); err != nil {
			n.logger.ErrorContext(ctx, "failed to scan unsent notification row", "err", err)
			continue
		}

		subject := fmt.Sprintf("【庫存警示】%s（SKU: %s）剩餘 %d 件", part.Name, part.SKU, part.Stock)
		body := fmt.Sprintf("零件名稱：%s\nSKU：%s\n目前庫存：%d\n再訂購水位：%d\n\n請盡快補貨。", part.Name, part.SKU, part.Stock, part.ReorderLevel)
		if err := n.mailer.Send(subject, body); err != nil {
			n.logger.ErrorContext(ctx, "retry: failed to send low stock email", "notification_id", id, "sku", part.SKU, "err", err)
			continue
		}
		_, _ = n.db.ExecContext(ctx, "UPDATE low_stock_notifications SET sent_at = ? WHERE id = ?", time.Now(), id)
		n.logger.InfoContext(ctx, "retry: low stock email sent", "notification_id", id, "sku", part.SKU)
	}
}
