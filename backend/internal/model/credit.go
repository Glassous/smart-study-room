package model

import "time"

// CreditLog 信用流水
type CreditLog struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Delta         int        `json:"delta"`
	Reason        string     `json:"reason"`
	ReservationID *int64     `json:"reservation_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreditOverview 信用概览(个人中心)
type CreditOverview struct {
	Score        int          `json:"score"`
	BannedUntil  *time.Time   `json:"banned_until"`
	Logs         []*CreditLog `json:"logs"`
}
