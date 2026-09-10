package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// CreditRepo 信用流水与信用分仓储
type CreditRepo struct {
	pool *pgxpool.Pool
}

func NewCreditRepo(pool *pgxpool.Pool) *CreditRepo {
	return &CreditRepo{pool: pool}
}

// ApplyDelta 事务: 变动信用分(夹紧0~100) + 写流水 + 低分自动禁约
// 返回(新分数, 是否触发禁约)
func (r *CreditRepo) ApplyDelta(ctx context.Context, userID int64, delta int, reason string, reservationID *int64) (int, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

	// 更新分数并回读
	var newScore int
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET credit_score = LEAST(100, GREATEST(0, credit_score + $2)),
		    updated_at = now()
		WHERE id = $1
		RETURNING credit_score`, userID, delta).Scan(&newScore)
	if err != nil {
		return 0, false, err
	}

	// 写流水
	if _, err := tx.Exec(ctx, `
		INSERT INTO credit_logs (user_id, delta, reason, reservation_id)
		VALUES ($1, $2, $3, $4)`,
		userID, delta, reason, reservationID); err != nil {
		return 0, false, err
	}

	// 低分禁约 / 恢复
	banned := false
	if newScore < 60 {
		if _, err := tx.Exec(ctx, `
			UPDATE users SET credit_banned_until = now() + interval '3 days'
			WHERE id = $1 AND credit_banned_until IS NULL`, userID); err != nil {
			return 0, false, err
		}
		banned = true
	} else if delta > 0 {
		// 分数回到阈值之上且禁约已过期时间在将来, 给予提前解除(奖励性恢复)
		if _, err := tx.Exec(ctx, `
			UPDATE users SET credit_banned_until = NULL
			WHERE id = $1 AND credit_banned_until IS NOT NULL AND credit_banned_until < now()`, userID); err != nil {
			return 0, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, false, err
	}
	return newScore, banned, nil
}

// ListByUser 用户信用流水(倒序)
func (r *CreditRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]*model.CreditLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, delta, reason, reservation_id, created_at
		FROM credit_logs WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.CreditLog
	for rows.Next() {
		var l model.CreditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Delta, &l.Reason, &l.ReservationID, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &l)
	}
	return list, rows.Err()
}
