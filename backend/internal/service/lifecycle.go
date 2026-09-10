package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// 生命周期规则参数
const (
	CheckinEarlyMinutes  = 15 // 最早可提前签到(分钟)
	CheckinGraceMinutes  = 15 // 开始后签到宽限(分钟), 超过判违约
	TempLeaveMaxMinutes  = 30 // 临时离开最长时间(分钟)
)

var (
	ErrNotInCheckinWindow = errors.New("不在签到时间窗口内(开始前后15分钟)")
	ErrLeaveTimeout       = errors.New("临时离开已超时")
)

// LifecycleHooks 生命周期事件钩子(通知/信用/候补在后续模块注入)
type LifecycleHooks interface {
	OnViolation(res *model.Reservation)
	OnComplete(res *model.Reservation)
}

type noopHooks struct{}

func (noopHooks) OnViolation(*model.Reservation) {}
func (noopHooks) OnComplete(*model.Reservation)  {}

// CompositeHooks 组合多个生命周期钩子(如 信用 + 通知)
type CompositeHooks struct {
	hooks []LifecycleHooks
}

func NewCompositeHooks(hs ...LifecycleHooks) *CompositeHooks {
	return &CompositeHooks{hooks: hs}
}

func (c *CompositeHooks) OnViolation(res *model.Reservation) {
	for _, h := range c.hooks {
		h.OnViolation(res)
	}
}

func (c *CompositeHooks) OnComplete(res *model.Reservation) {
	for _, h := range c.hooks {
		h.OnComplete(res)
	}
}

// LifecycleService 预约生命周期流转
type LifecycleService struct {
	reservations *repository.ReservationRepo
	hooks        LifecycleHooks
}

func NewLifecycleService(reservations *repository.ReservationRepo) *LifecycleService {
	return &LifecycleService{reservations: reservations, hooks: noopHooks{}}
}

// SetHooks 注入事件钩子
func (s *LifecycleService) SetHooks(h LifecycleHooks) {
	if h != nil {
		s.hooks = h
	}
}

// resStart 解析预约开始时间
func resStart(res *model.Reservation) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05",
		res.ResDate+" "+res.StartTime, time.Local)
}

func resEnd(res *model.Reservation) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05",
		res.ResDate+" "+res.EndTime, time.Local)
}

// Checkin 签到: pending → checked_in(窗口: 开始前15分钟 ~ 开始后15分钟)
func (s *LifecycleService) Checkin(ctx context.Context, userID, resID int64) error {
	res, err := s.getOwned(ctx, userID, resID)
	if err != nil {
		return err
	}
	if res.Status != model.ResPending {
		return ErrInvalidState
	}
	st, err := resStart(res)
	if err != nil {
		return ErrBadSlot
	}
	now := time.Now()
	if now.Before(st.Add(-CheckinEarlyMinutes*time.Minute)) ||
		now.After(st.Add(CheckinGraceMinutes*time.Minute)) {
		return ErrNotInCheckinWindow
	}
	updated, err := s.reservations.UpdateStatus(ctx, resID,
		[]string{model.ResPending}, model.ResCheckedIn,
		map[string]any{"checkin_at": now})
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidState
	}
	return nil
}

// Leave 临时离开: checked_in → temp_leave
func (s *LifecycleService) Leave(ctx context.Context, userID, resID int64) error {
	res, err := s.getOwned(ctx, userID, resID)
	if err != nil {
		return err
	}
	if res.Status != model.ResCheckedIn {
		return ErrInvalidState
	}
	updated, err := s.reservations.UpdateStatus(ctx, resID,
		[]string{model.ResCheckedIn}, model.ResTempLeave,
		map[string]any{"leave_at": time.Now()})
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidState
	}
	return nil
}

// ReturnBack 返回: temp_leave → checked_in(30 分钟内)
func (s *LifecycleService) ReturnBack(ctx context.Context, userID, resID int64) error {
	res, err := s.getOwned(ctx, userID, resID)
	if err != nil {
		return err
	}
	if res.Status != model.ResTempLeave {
		return ErrInvalidState
	}
	if res.LeaveAt != nil && time.Since(*res.LeaveAt) > TempLeaveMaxMinutes*time.Minute {
		return ErrLeaveTimeout
	}
	updated, err := s.reservations.UpdateStatus(ctx, resID,
		[]string{model.ResTempLeave}, model.ResCheckedIn, nil)
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidState
	}
	return nil
}

// Checkout 签退: checked_in → completed(提前结束)
func (s *LifecycleService) Checkout(ctx context.Context, userID, resID int64) error {
	res, err := s.getOwned(ctx, userID, resID)
	if err != nil {
		return err
	}
	if res.Status != model.ResCheckedIn && res.Status != model.ResTempLeave {
		return ErrInvalidState
	}
	updated, err := s.reservations.UpdateStatus(ctx, resID,
		[]string{model.ResCheckedIn, model.ResTempLeave}, model.ResCompleted,
		map[string]any{"checkout_at": time.Now()})
	if err != nil {
		return err
	}
	if !updated {
		return ErrInvalidState
	}
	// 触发履约事件(信用加分等, 由钩子消费)
	if fresh, err := s.reservations.GetByID(ctx, resID); err == nil {
		s.hooks.OnComplete(fresh)
	}
	return nil
}

// Tick 调度器单轮: 违约扫描 / 到时自动完成 / 临时离开超时
func (s *LifecycleService) Tick(ctx context.Context) (violations, completes, leaveTimeouts int) {
	noShows, err := s.reservations.ExpireNoShows(ctx)
	if err != nil {
		log.Printf("[scheduler] 违约扫描失败: %v", err)
	}
	for _, res := range noShows {
		s.hooks.OnViolation(res)
	}

	done, err := s.reservations.AutoCompleteTimeout(ctx)
	if err != nil {
		log.Printf("[scheduler] 到时完成扫描失败: %v", err)
	}
	for _, res := range done {
		s.hooks.OnComplete(res)
	}

	leaveOut, err := s.reservations.EndTempLeaveTimeout(ctx)
	if err != nil {
		log.Printf("[scheduler] 临时离开超时扫描失败: %v", err)
	}
	for _, res := range leaveOut {
		s.hooks.OnComplete(res)
	}

	return len(noShows), len(done), len(leaveOut)
}

func (s *LifecycleService) getOwned(ctx context.Context, userID, resID int64) (*model.Reservation, error) {
	res, err := s.reservations.GetByID(ctx, resID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrResNotFound
		}
		return nil, err
	}
	if res.UserID != userID {
		return nil, ErrNotOwner
	}
	return res, nil
}
