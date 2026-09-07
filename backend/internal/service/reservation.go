package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// 预约业务规则参数
const (
	ResMaxDurationHours = 8  // 单次预约最长时长
	ResMinLeadMinutes   = 10 // 距开始最短提前量(分钟)
)

var (
	ErrResNotFound    = errors.New("预约单不存在")
	ErrNotOwner       = errors.New("无权操作他人的预约")
	ErrBadSlot        = errors.New("时段参数错误(需 HH:MM, 粒度30分钟, 且开始早于结束)")
	ErrOutOfOpenTime  = errors.New("预约时段超出房间开放时间")
	ErrTooLong        = errors.New("单次预约时长超出限制")
	ErrInPast         = errors.New("不能预约已开始或过期的时段")
	ErrSeatConflict   = errors.New("该座位此时段已被预约")
	ErrUserConflict   = errors.New("您在同一时段已有其他预约")
	ErrSeatUnavailable = errors.New("座位不存在或不可预约")
	ErrCreditBanned   = errors.New("信用分过低，暂时禁止预约")
	ErrInvalidState   = errors.New("当前状态不允许该操作")
)

// ReservationService 预约核心服务
type ReservationService struct {
	reservations *repository.ReservationRepo
	rooms        *repository.RoomRepo
	seats        *repository.SeatRepo
	users        *repository.UserRepo
}

func NewReservationService(
	reservations *repository.ReservationRepo,
	rooms *repository.RoomRepo,
	seats *repository.SeatRepo,
	users *repository.UserRepo,
) *ReservationService {
	return &ReservationService{reservations: reservations, rooms: rooms, seats: seats, users: users}
}

// validateSlot 校验时段格式/粒度/时长
func validateSlot(date, start, end string) (time.Time, time.Time, error) {
	st, err1 := time.ParseInLocation("2006-01-02 15:04", date+" "+start, time.Local)
	en, err2 := time.ParseInLocation("2006-01-02 15:04", date+" "+end, time.Local)
	if err1 != nil || err2 != nil || !en.After(st) {
		return time.Time{}, time.Time{}, ErrBadSlot
	}
	// 30 分钟粒度
	if st.Minute()%30 != 0 || en.Minute()%30 != 0 {
		return time.Time{}, time.Time{}, ErrBadSlot
	}
	if en.Sub(st) > ResMaxDurationHours*time.Hour {
		return time.Time{}, time.Time{}, ErrTooLong
	}
	return st, en, nil
}

// Create 创建预约(手动选座)
func (s *ReservationService) Create(ctx context.Context, userID int64, req *model.CreateReservationRequest) (*model.ReservationView, error) {
	st, _, err := validateSlot(req.Date, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	// 不能预约过去
	if st.Add(-ResMinLeadMinutes * time.Minute).Before(time.Now()) {
		return nil, ErrInPast
	}

	// 座位与房间
	seat, err := s.seats.GetByID(ctx, req.SeatID)
	if err != nil {
		return nil, ErrSeatUnavailable
	}
	if seat.Status != model.SeatAvailable {
		return nil, ErrSeatUnavailable
	}
	room, err := s.rooms.GetByID(ctx, seat.RoomID)
	if err != nil {
		return nil, ErrSeatUnavailable
	}
	// 开放时间校验(补齐秒后字符串比较即可, HH:MM:SS 字典序即时间序)
	if padTime(req.StartTime) < padTime(room.OpenTime) || padTime(req.EndTime) > padTime(room.CloseTime) {
		return nil, ErrOutOfOpenTime
	}

	// 信用禁约
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.CreditBannedUntil != nil && user.CreditBannedUntil.After(time.Now()) {
		return nil, ErrCreditBanned
	}

	// 应用层冲突预检(数据库排除约束兜底)
	seatConflict, userConflict, err := s.reservations.HasConflict(ctx, req.SeatID, userID,
		req.Date, padTime(req.StartTime), padTime(req.EndTime))
	if err != nil {
		return nil, err
	}
	if seatConflict {
		return nil, ErrSeatConflict
	}
	if userConflict {
		return nil, ErrUserConflict
	}

	res := &model.Reservation{
		UserID:    userID,
		SeatID:    req.SeatID,
		ResDate:   req.Date,
		StartTime: padTime(req.StartTime),
		EndTime:   padTime(req.EndTime),
		Status:    model.ResPending,
		Source:    model.SrcManual,
	}
	if err := s.reservations.Create(ctx, res); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			// 排除约束兜底: 判定具体冲突维度
			seatConflict, userConflict, _ := s.reservations.HasConflict(ctx, req.SeatID, userID,
				req.Date, padTime(req.StartTime), padTime(req.EndTime))
			if seatConflict {
				return nil, ErrSeatConflict
			}
			if userConflict {
				return nil, ErrUserConflict
			}
			return nil, ErrSeatConflict
		}
		return nil, err
	}
	return s.reservations.GetView(ctx, res.ID)
}

// Cancel 取消预约(pending 状态)
func (s *ReservationService) Cancel(ctx context.Context, userID, resID int64) error {
	res, err := s.reservations.GetByID(ctx, resID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrResNotFound
		}
		return err
	}
	if res.UserID != userID {
		return ErrNotOwner
	}
	if res.Status != model.ResPending {
		return ErrInvalidState
	}
	updated, err := s.reservations.UpdateStatus(ctx, resID,
		[]string{model.ResPending}, model.ResCancelled, nil)
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidState
	}
	return nil
}

// ListMine 我的预约
func (s *ReservationService) ListMine(ctx context.Context, userID int64) ([]*model.ReservationView, error) {
	return s.reservations.ListByUser(ctx, userID, 100)
}

// GetView 预约详情
func (s *ReservationService) GetView(ctx context.Context, id int64) (*model.ReservationView, error) {
	v, err := s.reservations.GetView(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrResNotFound
	}
	return v, err
}
