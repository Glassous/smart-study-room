package service

import (
	"context"
	"fmt"
	"log"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// 信用规则参数(与需求文档一致)
const (
	CreditViolationDelta = -8 // 超时未签到违约
	CreditLateCancelDelta = -2 // 开始前30分钟内取消
	CreditCompletionDelta = 1  // 按期签退履约
	CreditBanThreshold   = 60 // 低于该分数禁止预约
	CreditBanDays        = 3  // 禁约天数
)

// CreditService 信用分服务(实现 LifecycleHooks 自动联动)
type CreditService struct {
	credits *repository.CreditRepo
	users   *repository.UserRepo
}

func NewCreditService(credits *repository.CreditRepo, users *repository.UserRepo) *CreditService {
	return &CreditService{credits: credits, users: users}
}

// OnViolation 违约: 扣 8 分(LifecycleHooks)
func (s *CreditService) OnViolation(res *model.Reservation) {
	_, _, err := s.credits.ApplyDelta(context.Background(), res.UserID, CreditViolationDelta,
		fmt.Sprintf("预约 %s %s-%s 超时未签到，系统判定违约", res.ResDate, res.StartTime, res.EndTime),
		&res.ID)
	if err != nil {
		log.Printf("[credit] 违约扣分失败 res=%d: %v", res.ID, err)
	}
}

// OnComplete 履约: 加 1 分(LifecycleHooks)
func (s *CreditService) OnComplete(res *model.Reservation) {
	_, _, err := s.credits.ApplyDelta(context.Background(), res.UserID, CreditCompletionDelta,
		fmt.Sprintf("预约 %s %s-%s 履约完成", res.ResDate, res.StartTime, res.EndTime),
		&res.ID)
	if err != nil {
		log.Printf("[credit] 履约加分失败 res=%d: %v", res.ID, err)
	}
}

// ApplyLateCancel 迟到取消: 开始前30分钟内取消扣 2 分
func (s *CreditService) ApplyLateCancel(ctx context.Context, res *model.Reservation) error {
	_, _, err := s.credits.ApplyDelta(ctx, res.UserID, CreditLateCancelDelta,
		fmt.Sprintf("预约 %s %s 开始前30分钟内取消", res.ResDate, res.StartTime),
		&res.ID)
	return err
}

// GetOverview 信用概览(分数/禁约/流水)
func (s *CreditService) GetOverview(ctx context.Context, userID int64) (*model.CreditOverview, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	logs, err := s.credits.ListByUser(ctx, userID, 50)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []*model.CreditLog{}
	}
	return &model.CreditOverview{
		Score:       user.CreditScore,
		BannedUntil: user.CreditBannedUntil,
		Logs:        logs,
	}, nil
}
