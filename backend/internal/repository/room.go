package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// RoomRepo 自习室表仓储
type RoomRepo struct {
	pool *pgxpool.Pool
}

func NewRoomRepo(pool *pgxpool.Pool) *RoomRepo {
	return &RoomRepo{pool: pool}
}

const roomCols = `id, name, location, open_time::text, close_time::text,
	seat_rows, seat_cols, description, created_at, updated_at`

func scanRoom(row pgx.Row) (*model.Room, error) {
	var r model.Room
	err := row.Scan(&r.ID, &r.Name, &r.Location, &r.OpenTime, &r.CloseTime,
		&r.SeatRows, &r.SeatCols, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// List 房间列表(按 ID 升序)
func (r *RoomRepo) List(ctx context.Context) ([]*model.Room, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+roomCols+` FROM rooms ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Room
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, room)
	}
	return list, rows.Err()
}

// GetByID 按 ID 查询
func (r *RoomRepo) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	return scanRoom(r.pool.QueryRow(ctx,
		`SELECT `+roomCols+` FROM rooms WHERE id = $1`, id))
}

// Create 创建房间
func (r *RoomRepo) Create(ctx context.Context, room *model.Room) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO rooms (name, location, open_time, close_time, seat_rows, seat_cols, description)
		VALUES ($1, $2, $3::time, $4::time, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		room.Name, room.Location, room.OpenTime, room.CloseTime,
		room.SeatRows, room.SeatCols, room.Description).
		Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
}

// Update 更新房间
func (r *RoomRepo) Update(ctx context.Context, id int64, room *model.Room) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE rooms SET name=$2, location=$3, open_time=$4::time, close_time=$5::time,
			seat_rows=$6, seat_cols=$7, description=$8, updated_at=now()
		WHERE id = $1`,
		id, room.Name, room.Location, room.OpenTime, room.CloseTime,
		room.SeatRows, room.SeatCols, room.Description)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除房间(级联删除座位与预约)
func (r *RoomRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
