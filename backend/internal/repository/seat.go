package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// SeatRepo 座位表仓储
type SeatRepo struct {
	pool *pgxpool.Pool
}

func NewSeatRepo(pool *pgxpool.Pool) *SeatRepo {
	return &SeatRepo{pool: pool}
}

const seatCols = `id, room_id, seat_no, row_no, col_no, zone, has_power, near_window,
	status, created_at, updated_at`

func scanSeat(row pgx.Row) (*model.Seat, error) {
	var s model.Seat
	err := row.Scan(&s.ID, &s.RoomID, &s.SeatNo, &s.RowNo, &s.ColNo, &s.Zone,
		&s.HasPower, &s.NearWindow, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListByRoom 房间内座位(按行列排序)
func (r *SeatRepo) ListByRoom(ctx context.Context, roomID int64) ([]*model.Seat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+seatCols+` FROM seats WHERE room_id = $1 ORDER BY row_no, col_no`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Seat
	for rows.Next() {
		s, err := scanSeat(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// ListByRoomWithOccupancy 平面图查询: 附带目标时段占用状态
func (r *SeatRepo) ListByRoomWithOccupancy(ctx context.Context, roomID int64, date, start, end string) ([]*model.SeatWithOccupancy, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.room_id, s.seat_no, s.row_no, s.col_no, s.zone,
		       s.has_power, s.near_window, s.status, s.created_at, s.updated_at,
		       EXISTS (
		           SELECT 1 FROM reservations rv
		           WHERE rv.seat_id = s.id
		             AND rv.res_date = $2::date
		             AND rv.status IN ('pending', 'checked_in', 'temp_leave')
		             AND tsrange((rv.res_date + rv.start_time)::timestamp,
		                         (rv.res_date + rv.end_time)::timestamp)
		                 && tsrange(($2::date + $3::time)::timestamp, ($2::date + $4::time)::timestamp)
		       ) AS occupied
		FROM seats s
		WHERE s.room_id = $1
		ORDER BY s.row_no, s.col_no`, roomID, date, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.SeatWithOccupancy
	for rows.Next() {
		var item model.SeatWithOccupancy
		err := rows.Scan(&item.ID, &item.RoomID, &item.SeatNo, &item.RowNo, &item.ColNo,
			&item.Zone, &item.HasPower, &item.NearWindow, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.Occupied)
		if err != nil {
			return nil, err
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

// GetByID 按 ID 查询
func (r *SeatRepo) GetByID(ctx context.Context, id int64) (*model.Seat, error) {
	return scanSeat(r.pool.QueryRow(ctx,
		`SELECT `+seatCols+` FROM seats WHERE id = $1`, id))
}

// BatchCreate 按行列批量生成座位(事务内逐行插入)
func (r *SeatRepo) BatchCreate(ctx context.Context, roomID int64, seats []*model.Seat) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, s := range seats {
		_, err := tx.Exec(ctx, `
			INSERT INTO seats (room_id, seat_no, row_no, col_no, zone, has_power, near_window)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			roomID, s.SeatNo, s.RowNo, s.ColNo, s.Zone, s.HasPower, s.NearWindow)
		if err != nil {
			return fmt.Errorf("插入座位 %s 失败: %w", s.SeatNo, err)
		}
	}
	return tx.Commit(ctx)
}

// Update 更新座位属性/状态
func (r *SeatRepo) Update(ctx context.Context, id int64, req *model.SeatUpsertRequest) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE seats SET
			zone = $2,
			has_power = COALESCE($3, has_power),
			near_window = COALESCE($4, near_window),
			status = COALESCE(NULLIF($5, ''), status),
			updated_at = now()
		WHERE id = $1`,
		id, req.Zone, req.HasPower, req.NearWindow, req.Status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除座位
func (r *SeatRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM seats WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CountByRoom 房间座位数
func (r *SeatRepo) CountByRoom(ctx context.Context, roomID int64) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM seats WHERE room_id = $1`, roomID).Scan(&n)
	return n, err
}
