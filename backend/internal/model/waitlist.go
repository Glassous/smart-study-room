package model

import "time"

// WaitlistStatus 候补状态
const (
	WaitWaiting  = "waiting"
	WaitPromoted = "promoted"
	WaitCancelled = "cancelled"
	WaitExpired  = "expired"
)

// WaitlistEntry 候补记录
type WaitlistEntry struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	RoomID     int64     `json:"room_id"`
	ResDate    string    `json:"res_date"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	Zone       *string   `json:"zone"`
	HasPower   *bool     `json:"has_power"`
	NearWindow *bool     `json:"near_window"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WaitlistView 候补视图(含房间名/排队位置)
type WaitlistView struct {
	WaitlistEntry
	RoomName string `json:"room_name"`
	Position int    `json:"position"` // 同时段排队位次(仅 waiting 有效)
}

// JoinWaitlistRequest 加入候补请求
type JoinWaitlistRequest struct {
	RoomID     int64  `json:"room_id" binding:"required"`
	Date       string `json:"date" binding:"required,len=10"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	Zone       string `json:"zone" binding:"omitempty,oneof=quiet regular discussion computer"`
	HasPower   *bool  `json:"has_power"`
	NearWindow *bool  `json:"near_window"`
}
