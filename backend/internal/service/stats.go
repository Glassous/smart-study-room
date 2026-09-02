package service

import (
	"context"
	"errors"
	"time"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

var ErrStatsBadArg = errors.New("统计参数错误")

// StatsService 统计分析服务
type StatsService struct {
	stats *repository.StatsRepo
	seats *repository.SeatRepo
	rooms *repository.RoomRepo
}

func NewStatsService(stats *repository.StatsRepo, seats *repository.SeatRepo, rooms *repository.RoomRepo) *StatsService {
	return &StatsService{stats: stats, seats: seats, rooms: rooms}
}

// Heatmap 座位热力图(房间×日期)
func (s *StatsService) Heatmap(ctx context.Context, roomID int64, date string) (*model.HeatmapResponse, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrStatsBadArg
	}
	room, err := s.rooms.GetByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	seats, err := s.seats.ListByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	seatIDs, hours, flags, err := s.stats.SeatHourOccupancy(ctx, roomID, date)
	if err != nil {
		return nil, err
	}

	// 展开为 座位×小时 矩阵
	nSeats := len(seats)
	// 小时去重(保持顺序)
	var hourList []string
	hourIdx := map[string]int{}
	for _, h := range hours {
		if _, ok := hourIdx[h]; !ok {
			hourIdx[h] = len(hourList)
			hourList = append(hourList, h)
		}
	}
	nHours := len(hourList)

	cells := make([][]int, nSeats)
	for i := range cells {
		cells[i] = make([]int, nHours)
	}
	seatIdx := map[int64]int{}
	for i, st := range seats {
		seatIdx[st.ID] = i
	}
	// 平铺结果按 (seat, hour) 顺序返回
	for k := range seatIDs {
		si, ok1 := seatIdx[seatIDs[k]]
		hi, ok2 := hourIdx[hours[k]]
		if ok1 && ok2 && flags[k] {
			cells[si][hi] = 1
		}
	}

	// 每小时利用率
	hourly := make([]float64, nHours)
	for _, row := range cells {
		for j, v := range row {
			if v == 1 {
				hourly[j]++
			}
		}
	}
	if nSeats > 0 {
		for j := range hourly {
			hourly[j] /= float64(nSeats)
		}
	}

	return &model.HeatmapResponse{
		Room:        room,
		Hours:       hourList,
		Seats:       seats,
		Cells:       cells,
		HourlyUsage: hourly,
	}, nil
}

// Trend 近 N 天使用率趋势
func (s *StatsService) Trend(ctx context.Context, days int) ([]*model.TrendPoint, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	dates, utils, counts, err := s.stats.UtilizationByDate(ctx, days)
	if err != nil {
		return nil, err
	}
	points := make([]*model.TrendPoint, 0, len(dates))
	for i := range dates {
		points = append(points, &model.TrendPoint{Date: dates[i], Utilization: utils[i], Count: counts[i]})
	}
	return points, nil
}

// Peak 近 N 天高峰时段分布
func (s *StatsService) Peak(ctx context.Context, days int) ([]*model.PeakPoint, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	hours, counts, err := s.stats.PeakHours(ctx, days)
	if err != nil {
		return nil, err
	}
	points := make([]*model.PeakPoint, 0, len(hours))
	for i := range hours {
		points = append(points, &model.PeakPoint{Hour: hours[i], Count: counts[i]})
	}
	return points, nil
}

// TopSeats 近 N 天热门座位
func (s *StatsService) TopSeats(ctx context.Context, days, limit int) ([]*model.TopSeat, error) {
	if days <= 0 || days > 90 {
		days = 14
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	seatNos, roomNames, hours, counts, err := s.stats.TopSeats(ctx, days, limit)
	if err != nil {
		return nil, err
	}
	list := make([]*model.TopSeat, 0, len(seatNos))
	for i := range seatNos {
		list = append(list, &model.TopSeat{
			SeatNo: seatNos[i], RoomName: roomNames[i], Hours: hours[i], Count: counts[i],
		})
	}
	return list, nil
}

// Overview 运营总览(今日 + 趋势 + 高峰 + 热门)
func (s *StatsService) Overview(ctx context.Context) (*model.StatsOverview, error) {
	totalSeats, activeToday, inUseNow, usedHours, capHours, err := s.stats.TodaySummary(ctx)
	if err != nil {
		return nil, err
	}
	util := 0.0
	if capHours > 0 {
		util = usedHours / capHours
	}
	trend, err := s.Trend(ctx, 14)
	if err != nil {
		return nil, err
	}
	peak, err := s.Peak(ctx, 14)
	if err != nil {
		return nil, err
	}
	top, err := s.TopSeats(ctx, 14, 10)
	if err != nil {
		return nil, err
	}
	return &model.StatsOverview{
		TotalSeats:       totalSeats,
		ActiveToday:      activeToday,
		InUseNow:         inUseNow,
		TodayUtilization: util,
		Trend14:          trend,
		Peak:             peak,
		TopSeats:         top,
	}, nil
}
