package model

// HeatmapResponse 座位热力图
// cells[i][j] = 座位 i 在小时 j 是否被占用(1/0)
type HeatmapResponse struct {
	Room        *Room    `json:"room"`
	Hours       []string `json:"hours"`        // "08:00".."21:00"
	Seats       []*Seat  `json:"seats"`        // 行列排序
	Cells       [][]int  `json:"cells"`        // 座位×小时占用矩阵
	HourlyUsage []float64 `json:"hourly_usage"` // 每小时房间利用率(0~1)
}

// TrendPoint 使用率趋势点
type TrendPoint struct {
	Date        string  `json:"date"`
	Utilization float64 `json:"utilization"`
	Count       int     `json:"count"`
}

// PeakPoint 高峰时段分布点
type PeakPoint struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

// TopSeat 热门座位
type TopSeat struct {
	SeatNo   string  `json:"seat_no"`
	RoomName string  `json:"room_name"`
	Hours    float64 `json:"hours"`
	Count    int     `json:"count"`
}

// StatsOverview 运营统计总览
type StatsOverview struct {
	TotalSeats      int         `json:"total_seats"`
	ActiveToday     int         `json:"active_today"`
	InUseNow        int         `json:"in_use_now"`
	TodayUtilization float64    `json:"today_utilization"`
	Trend14         []*TrendPoint `json:"trend_14"`
	Peak            []*PeakPoint  `json:"peak"`
	TopSeats        []*TopSeat    `json:"top_seats"`
}
