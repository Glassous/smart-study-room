package model

import "time"

// Zone 座位功能区域
const (
	ZoneQuiet     = "quiet"
	ZoneRegular   = "regular"
	ZoneDiscussion = "discussion"
	ZoneComputer  = "computer"
)

// SeatStatus 座位状态
const (
	SeatAvailable  = "available"
	SeatMaintenance = "maintenance"
	SeatDisabled   = "disabled"
)

// Room 自习室
type Room struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	OpenTime    string    `json:"open_time"`
	CloseTime   string    `json:"close_time"`
	SeatRows    int       `json:"seat_rows"`
	SeatCols    int       `json:"seat_cols"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoomUpsertRequest 创建/更新自习室请求
type RoomUpsertRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Location    string `json:"location" binding:"max=200"`
	OpenTime    string `json:"open_time" binding:"required"`
	CloseTime   string `json:"close_time" binding:"required"`
	SeatRows    int    `json:"seat_rows" binding:"min=1,max=30"`
	SeatCols    int    `json:"seat_cols" binding:"min=1,max=30"`
	Description string `json:"description"`
}

// Seat 座位
type Seat struct {
	ID         int64     `json:"id"`
	RoomID     int64     `json:"room_id"`
	SeatNo     string    `json:"seat_no"`
	RowNo      int       `json:"row_no"`
	ColNo      int       `json:"col_no"`
	Zone       string    `json:"zone"`
	HasPower   bool      `json:"has_power"`
	NearWindow bool      `json:"near_window"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SeatWithOccupancy 座位平面图条目(含目标时段占用状态)
type SeatWithOccupancy struct {
	Seat
	Occupied bool `json:"occupied"`
}

// SeatMapResponse 座位平面图响应
type SeatMapResponse struct {
	Room *Room               `json:"room"`
	Seats []*SeatWithOccupancy `json:"seats"`
}

// SeatUpsertRequest 更新座位请求
type SeatUpsertRequest struct {
	Zone       string `json:"zone" binding:"required,oneof=quiet regular discussion computer"`
	HasPower   *bool  `json:"has_power"`
	NearWindow *bool  `json:"near_window"`
	Status     string `json:"status" binding:"omitempty,oneof=available maintenance disabled"`
}

// BatchGenSeatsRequest 批量生成座位请求
type BatchGenSeatsRequest struct {
	SeatRows int `json:"seat_rows" binding:"min=1,max=30"`
	SeatCols int `json:"seat_cols" binding:"min=1,max=30"`
}
