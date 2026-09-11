package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// NotificationRepo 通知表仓储
type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

// Insert 写入通知(业务失败不阻断主流程, 由调用方降级处理)
func (r *NotificationRepo) Insert(ctx context.Context, n *model.Notification) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, type, title, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_read, created_at`,
		n.UserID, n.Type, n.Title, n.Content).
		Scan(&n.ID, &n.IsRead, &n.CreatedAt)
}

// ListByUser 用户通知(倒序)
func (r *NotificationRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]*model.Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, type, title, content, is_read, created_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &n)
	}
	return list, rows.Err()
}

// CountUnread 未读数
func (r *NotificationRepo) CountUnread(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM notifications WHERE user_id = $1 AND NOT is_read`, userID).Scan(&n)
	return n, err
}

// MarkRead 标记单条已读(仅本人的)
func (r *NotificationRepo) MarkRead(ctx context.Context, userID, id int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkAllRead 全部已读
func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND NOT is_read`, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
