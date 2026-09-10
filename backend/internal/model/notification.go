package model

import "time"

// 通知类型
const (
	NotifyReservationSuccess = "reservation_success"
	NotifyCheckinReminder    = "checkin_reminder"
	NotifyViolation          = "violation"
	NotifyCreditChange       = "credit_change"
	NotifyWaitlistPromoted   = "waitlist_promoted"
	NotifySystem             = "system"
)

// Notification 站内通知
type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
