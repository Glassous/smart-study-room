package service

import (
	"context"
	"errors"
	"log"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// NotificationService 站内消息通知服务
type NotificationService struct {
	notifications *repository.NotificationRepo
}

func NewNotificationService(notifications *repository.NotificationRepo) *NotificationService {
	return &NotificationService{notifications: notifications}
}

// Push 写入通知(失败仅记日志, 不阻断主业务)
func (s *NotificationService) Push(ctx context.Context, userID int64, typ, title, content string) {
	n := &model.Notification{UserID: userID, Type: typ, Title: title, Content: content}
	if err := s.notifications.Insert(ctx, n); err != nil {
		log.Printf("[notify] 写入通知失败 user=%d type=%s: %v", userID, typ, err)
	}
}

// List 用户通知列表
func (s *NotificationService) List(ctx context.Context, userID int64) ([]*model.Notification, error) {
	list, err := s.notifications.ListByUser(ctx, userID, 100)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []*model.Notification{}
	}
	return list, nil
}

// UnreadCount 未读数
func (s *NotificationService) UnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.notifications.CountUnread(ctx, userID)
}

// MarkRead 标记已读
func (s *NotificationService) MarkRead(ctx context.Context, userID, id int64) error {
	err := s.notifications.MarkRead(ctx, userID, id)
	if errors.Is(err, repository.ErrNotFound) {
		return errors.New("通知不存在")
	}
	return err
}

// MarkAllRead 全部已读
func (s *NotificationService) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	return s.notifications.MarkAllRead(ctx, userID)
}

// ---- LifecycleHooks 实现: 违约即时警告 ----

// OnViolation 违约警告通知
func (s *NotificationService) OnViolation(res *model.Reservation) {
	s.Push(context.Background(), res.UserID, model.NotifyViolation,
		"预约违约警告",
		"您的预约("+res.ResDate+" "+res.StartTime[:5]+"-"+res.EndTime[:5]+
			")超时未签到已被判定违约，信用分 -8，请按时签到避免损失。")
}

// OnComplete 履约完成不发通知(避免打扰, 结果在个人中心可见)
func (s *NotificationService) OnComplete(*model.Reservation) {}
