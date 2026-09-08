package model

import "time"

// ReservationStatus 预约状态机
const (
	ResPending   = "pending"    // 待签到
	ResCheckedIn = "checked_in" // 使用中
	ResTempLeave = "temp_leave" // 临时离开
	ResCompleted = "completed"  // 已完成(签退/到时)
	ResCancelled = "cancelled"  // 已取消
	ResViolation = "violation"  // 违约(超时未签到)
)

// ActiveStatuses 占用座位的活跃状态
var ActiveStatuses = []string{ResPending, ResCheckedIn, ResTempLeave}

// ReservationSource 预约来源
const (
	SrcManual  = "manual"
	SrcAuto    = "auto"
	SrcWaitlist = "waitlist"
)

// Reservation 预约单
type Reservation struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	SeatID     int64      `json:"seat_id"`
	ResDate    string     `json:"res_date"`
	StartTime  string     `json:"start_time"`
	EndTime    string     `json:"end_time"`
	Status     string     `json:"status"`
	Source     string     `json:"source"`
	CheckinAt  *time.Time `json:"checkin_at"`
	CheckoutAt *time.Time `json:"checkout_at"`
	LeaveAt    *time.Time `json:"leave_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ReservationView 预约列表视图(联表座位/房间/用户)
type ReservationView struct {
	Reservation
	SeatNo   string `json:"seat_no"`
	Zone     string `json:"zone"`
	HasPower bool   `json:"has_power"`
	RoomID   int64  `json:"room_id"`
	RoomName string `json:"room_name"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
}

// CreateReservationRequest 创建预约请求
type CreateReservationRequest struct {
	SeatID    int64  `json:"seat_id" binding:"required"`
	Date      string `json:"date" binding:"required,len=10"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}
