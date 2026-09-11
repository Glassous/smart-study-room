package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// WaitlistRepo 候补表仓储
type WaitlistRepo struct {
	pool *pgxpool.Pool
}

func NewWaitlistRepo(pool *pgxpool.Pool) *WaitlistRepo {
	return &WaitlistRepo{pool: pool}
}

const waitCols = `id, user_id, room_id, res_date::text, start_time::text, end_time::text,
	zone, has_power, near_window, status, created_at, updated_at`

func scanWaitlist(row pgx.Row) (*model.WaitlistEntry, error) {
	var w model.WaitlistEntry
	err := row.Scan(&w.ID, &w.UserID, &w.RoomID, &w.ResDate, &w.StartTime, &w.EndTime,
		&w.Zone, &w.HasPower, &w.NearWindow, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// Insert 加入候补
func (r *WaitlistRepo) Insert(ctx context.Context, w *model.WaitlistEntry) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO waitlist (user_id, room_id, res_date, start_time, end_time, zone, has_power, near_window)
		VALUES ($1, $2, $3::date, $4::time, $5::time, NULLIF($6, ''), $7, $8)
		RETURNING id, created_at, updated_at`,
		w.UserID, w.RoomID, w.ResDate, w.StartTime, w.EndTime, derefStr(w.Zone), w.HasPower, w.NearWindow).
		Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

// GetByID 按 ID 查询
func (r *WaitlistRepo) GetByID(ctx context.Context, id int64) (*model.WaitlistEntry, error) {
	return scanWaitlist(r.pool.QueryRow(ctx,
		`SELECT `+waitCols+` FROM waitlist WHERE id = $1`, id))
}

// HasWaitingSameSlot 同一用户同一房间同一时段是否已在排队
func (r *WaitlistRepo) HasWaitingSameSlot(ctx context.Context, userID, roomID int64, date, start, end string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM waitlist
			WHERE user_id = $1 AND room_id = $2 AND res_date = $3::date
			  AND start_time = $4::time AND end_time = $5::time AND status = 'waiting')`,
		userID, roomID, date, start, end).Scan(&exists)
	return exists, err
}

// ListByUser 我的候补(含房间名与排队位次)
func (r *WaitlistRepo) ListByUser(ctx context.Context, userID int64) ([]*model.WaitlistView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT w.id, w.user_id, w.room_id, w.res_date::text, w.start_time::text, w.end_time::text,
		       w.zone, w.has_power, w.near_window, w.status, w.created_at, w.updated_at,
		       rm.name,
		       CASE WHEN w.status = 'waiting' THEN (
		           SELECT count(*) FROM waitlist w2
		           WHERE w2.room_id = w.room_id AND w2.res_date = w.res_date
		             AND w2.start_time = w.start_time AND w2.end_time = w.end_time
		             AND w2.status = 'waiting' AND w2.created_at <= w.created_at
		       ) ELSE 0 END AS position
		FROM waitlist w JOIN rooms rm ON rm.id = w.room_id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.WaitlistView
	for rows.Next() {
		var v model.WaitlistView
		if err := rows.Scan(&v.ID, &v.UserID, &v.RoomID, &v.ResDate, &v.StartTime, &v.EndTime,
			&v.Zone, &v.HasPower, &v.NearWindow, &v.Status, &v.CreatedAt, &v.UpdatedAt,
			&v.RoomName, &v.Position); err != nil {
			return nil, err
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

// ListWaitingSlots 去重列出当前仍在等待的 (房间,日期,时段) 组合(调度递补扫描用)
func (r *WaitlistRepo) ListWaitingSlots(ctx context.Context) ([]*model.WaitlistEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (room_id, res_date, start_time, end_time)
		       ` + waitCols + `
		FROM waitlist
		WHERE status = 'waiting' AND res_date >= CURRENT_DATE
		ORDER BY room_id, res_date, start_time, end_time, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.WaitlistEntry
	for rows.Next() {
		w, err := scanWaitlist(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

// ListQueue 某时段的等待队列(按排队顺序)
func (r *WaitlistRepo) ListQueue(ctx context.Context, roomID int64, date, start, end string, limit int) ([]*model.WaitlistEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+waitCols+` FROM waitlist
		WHERE room_id = $1 AND res_date = $2::date
		  AND start_time = $3::time AND end_time = $4::time AND status = 'waiting'
		ORDER BY created_at LIMIT $5`, roomID, date, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.WaitlistEntry
	for rows.Next() {
		w, err := scanWaitlist(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

// UpdateStatus 条件状态更新
func (r *WaitlistRepo) UpdateStatus(ctx context.Context, id int64, from, to string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE waitlist SET status = $3, updated_at = now()
		 WHERE id = $1 AND status = $2`, id, from, to)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ExpireOverdue 过期未递补的候补(日期已过) → expired
func (r *WaitlistRepo) ExpireOverdue(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE waitlist SET status = 'expired', updated_at = now()
		WHERE status = 'waiting' AND res_date < CURRENT_DATE`)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
