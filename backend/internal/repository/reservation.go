package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// ReservationRepo 预约表仓储
type ReservationRepo struct {
	pool *pgxpool.Pool
}

func NewReservationRepo(pool *pgxpool.Pool) *ReservationRepo {
	return &ReservationRepo{pool: pool}
}

const resCols = `id, user_id, seat_id, res_date::text, start_time::text, end_time::text,
	status, source, checkin_at, checkout_at, leave_at, created_at, updated_at`

const resViewCols = `r.id, r.user_id, r.seat_id, r.res_date::text, r.start_time::text, r.end_time::text,
	r.status, r.source, r.checkin_at, r.checkout_at, r.leave_at, r.created_at, r.updated_at,
	s.seat_no, s.zone, s.has_power, s.room_id, rm.name, u.username, u.real_name`

const resViewFrom = `FROM reservations r
	JOIN seats s ON s.id = r.seat_id
	JOIN rooms rm ON rm.id = s.room_id
	JOIN users u ON u.id = r.user_id`

func scanReservation(row pgx.Row) (*model.Reservation, error) {
	var r model.Reservation
	err := row.Scan(&r.ID, &r.UserID, &r.SeatID, &r.ResDate, &r.StartTime, &r.EndTime,
		&r.Status, &r.Source, &r.CheckinAt, &r.CheckoutAt, &r.LeaveAt,
		&r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func scanView(row pgx.Row) (*model.ReservationView, error) {
	var v model.ReservationView
	err := row.Scan(&v.ID, &v.UserID, &v.SeatID, &v.ResDate, &v.StartTime, &v.EndTime,
		&v.Status, &v.Source, &v.CheckinAt, &v.CheckoutAt, &v.LeaveAt,
		&v.CreatedAt, &v.UpdatedAt,
		&v.SeatNo, &v.Zone, &v.HasPower, &v.RoomID, &v.RoomName, &v.Username, &v.RealName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Create 创建预约(时段冲突由数据库排除约束兜底)
func (r *ReservationRepo) Create(ctx context.Context, res *model.Reservation) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO reservations (user_id, seat_id, res_date, start_time, end_time, status, source)
		VALUES ($1, $2, $3::date, $4::time, $5::time, $6, $7)
		RETURNING id, created_at, updated_at`,
		res.UserID, res.SeatID, res.ResDate, res.StartTime, res.EndTime, res.Status, res.Source).
		Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
}

// GetByID 按 ID 查询
func (r *ReservationRepo) GetByID(ctx context.Context, id int64) (*model.Reservation, error) {
	return scanReservation(r.pool.QueryRow(ctx,
		`SELECT `+resCols+` FROM reservations WHERE id = $1`, id))
}

// GetView 按 ID 查询联表视图
func (r *ReservationRepo) GetView(ctx context.Context, id int64) (*model.ReservationView, error) {
	return scanView(r.pool.QueryRow(ctx,
		`SELECT `+resViewCols+` `+resViewFrom+` WHERE r.id = $1`, id))
}

// ListByUser 用户预约列表(倒序)
func (r *ReservationRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]*model.ReservationView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+resViewCols+` `+resViewFrom+`
		WHERE r.user_id = $1
		ORDER BY r.res_date DESC, r.start_time DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.ReservationView
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

// ListAll 全量预约(管理端, 可按状态过滤)
func (r *ReservationRepo) ListAll(ctx context.Context, status string, limit int) ([]*model.ReservationView, error) {
	q := `SELECT ` + resViewCols + ` ` + resViewFrom
	var args []any
	if status != "" {
		args = append(args, status)
		q += fmt.Sprintf(` WHERE r.status = $%d`, len(args))
	}
	args = append(args, limit)
	q += fmt.Sprintf(` ORDER BY r.res_date DESC, r.start_time DESC LIMIT $%d`, len(args))
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.ReservationView
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

// HasConflict 检测目标时段的座位/用户冲突
// 占位符: $1=seatID $2=userID $3=date $4=start $5=end $6=活跃状态集
const conflictSQL = `
SELECT
	EXISTS(SELECT 1 FROM reservations rv
	       WHERE rv.seat_id = $1 AND rv.res_date = $3::date AND rv.status = ANY($6)
	         AND tsrange((rv.res_date + rv.start_time)::timestamp, (rv.res_date + rv.end_time)::timestamp)
	           && tsrange(($3::date + $4::time)::timestamp, ($3::date + $5::time)::timestamp)),
	EXISTS(SELECT 1 FROM reservations rv
	       WHERE rv.user_id = $2 AND rv.res_date = $3::date AND rv.status = ANY($6)
	         AND tsrange((rv.res_date + rv.start_time)::timestamp, (rv.res_date + rv.end_time)::timestamp)
	           && tsrange(($3::date + $4::time)::timestamp, ($3::date + $5::time)::timestamp))`

// HasConflict 返回(座位冲突, 用户冲突)
func (r *ReservationRepo) HasConflict(ctx context.Context, seatID, userID int64, date, start, end string) (bool, bool, error) {
	var seatConflict, userConflict bool
	err := r.pool.QueryRow(ctx, conflictSQL,
		seatID, userID, date, start, end, model.ActiveStatuses).
		Scan(&seatConflict, &userConflict)
	return seatConflict, userConflict, err
}

// FindUserReservationInSlot 查找用户在指定时段的有效预约(用于判断是否已在该时段选座)
func (r *ReservationRepo) FindUserReservationInSlot(ctx context.Context, userID int64, date, start, end string) (*model.Reservation, error) {
	var res model.Reservation
	statuses := []string{model.ResPending, model.ResCheckedIn, model.ResTempLeave, model.ResCompleted}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, seat_id, res_date::text, start_time::text, end_time::text, status
		FROM reservations
		WHERE user_id = $1 AND res_date = $2::date
		  AND status = ANY($5)
		  AND tsrange((res_date + start_time)::timestamp, (res_date + end_time)::timestamp)
		   && tsrange(($2::date + $3::time)::timestamp, ($2::date + $4::time)::timestamp)
		LIMIT 1`,
		userID, date, start, end, statuses).
		Scan(&res.ID, &res.UserID, &res.SeatID, &res.ResDate, &res.StartTime, &res.EndTime, &res.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateStatus 条件状态迁移(仅当当前状态在 from 集合内才更新), 返回是否生效
func (r *ReservationRepo) UpdateStatus(ctx context.Context, id int64, from []string, to string, sets map[string]any) (bool, error) {
	q := `UPDATE reservations SET status = $2, updated_at = now()`
	args := []any{id, to}
	// 附加更新列(如 checkin_at = now())
	for col, val := range sets {
		args = append(args, val)
		q += fmt.Sprintf(`, %s = $%d`, col, len(args))
	}
	args = append(args, from)
	q += fmt.Sprintf(` WHERE id = $1 AND status = ANY($%d)`, len(args))
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ---- 调度器批量状态迁移(原子 UPDATE ... RETURNING) ----

// ExpireNoShows 超时未签到 → 违约(开始后15分钟仍未签到)
// 仅处理近两日数据, 避免历史数据被翻动
func (r *ReservationRepo) ExpireNoShows(ctx context.Context) ([]*model.Reservation, error) {
	return r.batchUpdate(ctx, `
		UPDATE reservations SET status = 'violation', updated_at = now()
		WHERE status = 'pending'
		  AND res_date >= CURRENT_DATE - 1
		  AND (res_date + start_time)::timestamp < now() - interval '15 minutes'
		RETURNING `+resCols)
}

// AutoCompleteTimeout 到时未签退 → 已完成(使用中/临时离开且已过结束时间)
func (r *ReservationRepo) AutoCompleteTimeout(ctx context.Context) ([]*model.Reservation, error) {
	return r.batchUpdate(ctx, `
		UPDATE reservations
		SET status = 'completed', checkout_at = (res_date + end_time), updated_at = now()
		WHERE status IN ('checked_in', 'temp_leave')
		  AND res_date >= CURRENT_DATE - 1
		  AND (res_date + end_time)::timestamp < now()
		RETURNING `+resCols)
}

// EndTempLeaveTimeout 临时离开超时(30分钟) → 已完成
func (r *ReservationRepo) EndTempLeaveTimeout(ctx context.Context) ([]*model.Reservation, error) {
	return r.batchUpdate(ctx, `
		UPDATE reservations
		SET status = 'completed', checkout_at = now(), updated_at = now()
		WHERE status = 'temp_leave'
		  AND res_date >= CURRENT_DATE - 1
		  AND leave_at < now() - interval '30 minutes'
		RETURNING `+resCols)
}

// EndOverdueTempLeave 临时离开超时返回失败(供 Return 判断): 不更新, 仅查询

func (r *ReservationRepo) batchUpdate(ctx context.Context, q string) ([]*model.Reservation, error) {
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Reservation
	for rows.Next() {
		res, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, res)
	}
	return list, rows.Err()
}
