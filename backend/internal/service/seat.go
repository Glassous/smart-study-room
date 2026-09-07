package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

var (
	ErrRoomNotFound  = errors.New("自习室不存在")
	ErrSeatNotFound  = errors.New("座位不存在")
	ErrRoomHasSeats  = errors.New("该房间已存在座位，请先清空座位再批量生成")
	ErrBadTimeRange  = errors.New("开放时间格式错误或结束时间不晚于开始时间")
	ErrBadSeatMapArg = errors.New("平面图查询参数错误(需要 date/start/end)")
)

var timeRe = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$|^\d{2}:\d{2}$`)

// SeatService 房间与座位管理服务
type SeatService struct {
	rooms *repository.RoomRepo
	seats *repository.SeatRepo
}

func NewSeatService(rooms *repository.RoomRepo, seats *repository.SeatRepo) *SeatService {
	return &SeatService{rooms: rooms, seats: seats}
}

// ListRooms 房间列表
func (s *SeatService) ListRooms(ctx context.Context) ([]*model.Room, error) {
	return s.rooms.List(ctx)
}

// GetRoom 房间详情
func (s *SeatService) GetRoom(ctx context.Context, id int64) (*model.Room, error) {
	r, err := s.rooms.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrRoomNotFound
	}
	return r, err
}

func validateTimeRange(open, close string) error {
	if !timeRe.MatchString(open) || !timeRe.MatchString(close) {
		return ErrBadTimeRange
	}
	o, err1 := time.Parse("15:04:05", padTime(open))
	c, err2 := time.Parse("15:04:05", padTime(close))
	if err1 != nil || err2 != nil || !c.After(o) {
		return ErrBadTimeRange
	}
	return nil
}

func padTime(t string) string {
	if len(t) == 5 {
		return t + ":00"
	}
	return t
}

// CreateRoom 创建房间
func (s *SeatService) CreateRoom(ctx context.Context, req *model.RoomUpsertRequest) (*model.Room, error) {
	if err := validateTimeRange(req.OpenTime, req.CloseTime); err != nil {
		return nil, err
	}
	room := &model.Room{
		Name: req.Name, Location: req.Location,
		OpenTime: padTime(req.OpenTime), CloseTime: padTime(req.CloseTime),
		SeatRows: req.SeatRows, SeatCols: req.SeatCols, Description: req.Description,
	}
	if err := s.rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

// UpdateRoom 更新房间
func (s *SeatService) UpdateRoom(ctx context.Context, id int64, req *model.RoomUpsertRequest) (*model.Room, error) {
	if err := validateTimeRange(req.OpenTime, req.CloseTime); err != nil {
		return nil, err
	}
	room := &model.Room{
		Name: req.Name, Location: req.Location,
		OpenTime: padTime(req.OpenTime), CloseTime: padTime(req.CloseTime),
		SeatRows: req.SeatRows, SeatCols: req.SeatCols, Description: req.Description,
	}
	if err := s.rooms.Update(ctx, id, room); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	return s.GetRoom(ctx, id)
}

// DeleteRoom 删除房间
func (s *SeatService) DeleteRoom(ctx context.Context, id int64) error {
	err := s.rooms.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrRoomNotFound
	}
	return err
}

// BatchGenSeats 按行列批量生成座位
// 属性规则(与演示种子一致): 最右列靠窗; 偶数列带电源;
// 区域按行分配: 前 1/3 行静音、中间普通、后段研讨
func (s *SeatService) BatchGenSeats(ctx context.Context, roomID int64, rows, cols int) (int, error) {
	if _, err := s.GetRoom(ctx, roomID); err != nil {
		return 0, err
	}
	n, err := s.seats.CountByRoom(ctx, roomID)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		return 0, ErrRoomHasSeats
	}

	quietRows := rows / 3
	if quietRows < 1 {
		quietRows = 1
	}
	discussionRows := rows - quietRows - rows/3
	if discussionRows < 0 {
		discussionRows = 0
	}

	seats := make([]*model.Seat, 0, rows*cols)
	for row := 1; row <= rows; row++ {
		zone := model.ZoneRegular
		switch {
		case row <= quietRows:
			zone = model.ZoneQuiet
		case row > rows-discussionRows && discussionRows > 0:
			zone = model.ZoneDiscussion
		}
		for col := 1; col <= cols; col++ {
			seats = append(seats, &model.Seat{
				RoomID:     roomID,
				SeatNo:     fmt.Sprintf("%c%02d", 'A'+row-1, col),
				RowNo:      row,
				ColNo:      col,
				Zone:       zone,
				HasPower:   col%2 == 0,
				NearWindow: col == cols,
			})
		}
	}
	if err := s.seats.BatchCreate(ctx, roomID, seats); err != nil {
		return 0, err
	}
	return len(seats), nil
}

// UpdateSeat 更新座位属性
func (s *SeatService) UpdateSeat(ctx context.Context, id int64, req *model.SeatUpsertRequest) error {
	err := s.seats.Update(ctx, id, req)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSeatNotFound
	}
	return err
}

// DeleteSeat 删除座位
func (s *SeatService) DeleteSeat(ctx context.Context, id int64) error {
	err := s.seats.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSeatNotFound
	}
	return err
}

// GetSeatMap 座位平面图(含目标时段占用)
func (s *SeatService) GetSeatMap(ctx context.Context, roomID int64, date, start, end string) (*model.SeatMapResponse, error) {
	room, err := s.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if date == "" || start == "" || end == "" {
		return nil, ErrBadSeatMapArg
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrBadSeatMapArg
	}
	if err := validateTimeRange(start, end); err != nil {
		return nil, ErrBadSeatMapArg
	}
	seats, err := s.seats.ListByRoomWithOccupancy(ctx, roomID, date, padTime(start), padTime(end))
	if err != nil {
		return nil, err
	}
	return &model.SeatMapResponse{Room: room, Seats: seats}, nil
}
