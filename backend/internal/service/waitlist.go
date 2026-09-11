package service

import (
	"context"
	"errors"
	"log"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

var (
	ErrWaitNotFound     = errors.New("候补记录不存在")
	ErrAlreadyWaiting   = errors.New("您已在同一场次候补队列中")
	ErrWaitInvalidState = errors.New("当前候补状态不允许该操作")
)

// PromoteTryCount 每个时段每轮最多尝试递补的队首人数
const PromoteTryCount = 3

// WaitlistService 满座候补服务
type WaitlistService struct {
	waitlist *repository.WaitlistRepo
	seats    *repository.SeatRepo
	res      *repository.ReservationRepo
	resSvc   *ReservationService
	notifier *NotificationService
}

func NewWaitlistService(
	waitlist *repository.WaitlistRepo,
	seats *repository.SeatRepo,
	res *repository.ReservationRepo,
	resSvc *ReservationService,
	notifier *NotificationService,
) *WaitlistService {
	return &WaitlistService{waitlist: waitlist, seats: seats, res: res, resSvc: resSvc, notifier: notifier}
}

// Join 加入候补队列
func (s *WaitlistService) Join(ctx context.Context, userID int64, req *model.JoinWaitlistRequest) (*model.WaitlistView, error) {
	if _, _, err := validateSlot(req.Date, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// 同场次重复排队
	exists, err := s.waitlist.HasWaitingSameSlot(ctx, userID, req.RoomID, req.Date,
		padTime(req.StartTime), padTime(req.EndTime))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyWaiting
	}

	entry := &model.WaitlistEntry{
		UserID:     userID,
		RoomID:     req.RoomID,
		ResDate:    req.Date,
		StartTime:  padTime(req.StartTime),
		EndTime:    padTime(req.EndTime),
		Status:     model.WaitWaiting,
	}
	if req.Zone != "" {
		entry.Zone = &req.Zone
	}
	entry.HasPower = req.HasPower
	entry.NearWindow = req.NearWindow

	if err := s.waitlist.Insert(ctx, entry); err != nil {
		return nil, err
	}
	list, err := s.waitlist.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.ID == entry.ID {
			return v, nil
		}
	}
	return &model.WaitlistView{WaitlistEntry: *entry}, nil
}

// Cancel 取消候补
func (s *WaitlistService) Cancel(ctx context.Context, userID, id int64) error {
	entry, err := s.waitlist.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrWaitNotFound
		}
		return err
	}
	if entry.UserID != userID {
		return ErrNotOwner
	}
	updated, err := s.waitlist.UpdateStatus(ctx, id, model.WaitWaiting, model.WaitCancelled)
	if err != nil {
		return err
	}
	if !updated {
		return ErrWaitInvalidState
	}
	return nil
}

// ListMine 我的候补
func (s *WaitlistService) ListMine(ctx context.Context, userID int64) ([]*model.WaitlistView, error) {
	list, err := s.waitlist.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*model.WaitlistView{}
	}
	return list, nil
}

// PromoteTick 调度递补: 扫描所有等待中的时段, 有空位即按队首递补
// 返回本轮成功递补数
func (s *WaitlistService) PromoteTick(ctx context.Context) int {
	slots, err := s.waitlist.ListWaitingSlots(ctx)
	if err != nil {
		log.Printf("[waitlist] 扫描候补时段失败: %v", err)
		return 0
	}

	promoted := 0
	for _, slot := range slots {
		// 该时段是否还有空位(复用自动分配候选集)
		candidates, err := s.seats.ListCandidates(ctx, slot.RoomID, slot.ResDate,
			slot.StartTime, slot.EndTime)
		if err != nil || len(candidates) == 0 {
			continue
		}

		queue, err := s.waitlist.ListQueue(ctx, slot.RoomID, slot.ResDate,
			slot.StartTime, slot.EndTime, PromoteTryCount)
		if err != nil {
			continue
		}
		// 依次尝试队首(最多 PromoteTryCount 人, 失败跳过)
		seats := candidates
		for _, entry := range queue {
			if len(seats) == 0 {
				break
			}
			seat := seats[0]
			view, err := s.resSvc.CreateWithSource(ctx, entry.UserID, &model.CreateReservationRequest{
				SeatID:    seat.ID,
				Date:      slot.ResDate,
				StartTime: slot.StartTime[:5],
				EndTime:   slot.EndTime[:5],
			}, model.SrcWaitlist)
			if err != nil {
				// 该用户冲突/禁约等, 跳过取下一位
				continue
			}
			if _, err := s.waitlist.UpdateStatus(ctx, entry.ID, model.WaitWaiting, model.WaitPromoted); err != nil {
				continue
			}
			promoted++
			s.notifier.Push(ctx, entry.UserID, model.NotifyWaitlistPromoted, "候补成功",
				"您候补的 "+view.RoomName+" "+view.SeatNo+"（"+slot.ResDate+" "+
					slot.StartTime[:5]+"-"+slot.EndTime[:5]+"）已递补成功，请按时签到。")
			// 递补占用一个空位
			seats = seats[1:]
		}
	}
	return promoted
}

// ExpireTick 过期候补清理
func (s *WaitlistService) ExpireTick(ctx context.Context) int {
	n, err := s.waitlist.ExpireOverdue(ctx)
	if err != nil {
		log.Printf("[waitlist] 过期清理失败: %v", err)
		return 0
	}
	return n
}
