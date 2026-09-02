package model

// AllocationRequest 自动分配请求
type AllocationRequest struct {
	RoomID     int64  `json:"room_id" binding:"required"`
	Date       string `json:"date" binding:"required,len=10"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	Zone       string `json:"zone" binding:"omitempty,oneof=quiet regular discussion computer"`
	HasPower   *bool  `json:"has_power"`
	NearWindow *bool  `json:"near_window"`
	TopN       int    `json:"top_n" binding:"omitempty,min=1,max=10"`
	AutoBook   bool   `json:"auto_book"`
}

// ScoredSeat 评分后的候选座位
type ScoredSeat struct {
	Seat    *Seat    `json:"seat"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

// AllocationResponse 自动分配响应
type AllocationResponse struct {
	Recommendations []*ScoredSeat      `json:"recommendations"`
	Booked          bool               `json:"booked"`
	Reservation     *ReservationView   `json:"reservation,omitempty"`
}

// CandidateSeat 候选座位(含近7日利用率, 由 SQL 计算填充)
type CandidateSeat struct {
	Seat
	Util7d float64 `json:"-"`
}
