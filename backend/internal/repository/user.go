package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

// ErrNotFound 通用未找到错误
var ErrNotFound = errors.New("record not found")

// UserRepo 用户表仓储
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

const userCols = `id, username, password_hash, real_name, student_no, role,
	credit_score, credit_banned_until, status, created_at, updated_at`

func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RealName, &u.StudentNo,
		&u.Role, &u.CreditScore, &u.CreditBannedUntil, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create 插入用户(回填默认值)
func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, real_name, student_no, role)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id, credit_score, status, created_at, updated_at`,
		u.Username, u.PasswordHash, u.RealName, nullableStr(u.StudentNo), u.Role).
		Scan(&u.ID, &u.CreditScore, &u.Status, &u.CreatedAt, &u.UpdatedAt)
}

// GetByUsername 按用户名查询
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE username = $1`, username))
}

// GetByID 按 ID 查询
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

// ExistsByUsername 判断用户名是否已存在
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

// ListAll 用户列表(管理端, 倒序, 简单上限)
func (r *UserRepo) ListAll(ctx context.Context, limit int) ([]*model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userCols+` FROM users ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RealName, &u.StudentNo,
			&u.Role, &u.CreditScore, &u.CreditBannedUntil, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// SetStatus 启用/禁用账号
func (r *UserRepo) SetStatus(ctx context.Context, id int64, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
