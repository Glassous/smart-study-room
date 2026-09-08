package service

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/imicola/smart-study-room/backend/internal/pkg/redissync"
)

// Scheduler 后台周期调度器
// 每轮依次执行: 违约扫描 → 到时自动完成 → 临时离开超时 → 候补递补 → 候补过期清理
type Scheduler struct {
	lifecycle *LifecycleService
	waitlist  *WaitlistService
	interval  time.Duration
	rdb       *redis.Client
}

func NewScheduler(lifecycle *LifecycleService, waitlist *WaitlistService, interval time.Duration) *Scheduler {
	return &Scheduler{lifecycle: lifecycle, waitlist: waitlist, interval: interval}
}

// SetRedis 设置 Redis 客户端以启用集群选主锁
func (s *Scheduler) SetRedis(rdb *redis.Client) {
	s.rdb = rdb
}

// Run 阻塞运行调度循环(由 main 以 goroutine 启动)
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	log.Printf("[scheduler] 启动, 周期 %s", s.interval)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[scheduler] 停止")
			return
		case <-ticker.C:
			func() {
				// 若启用 Redis，尝试抢占本轮分布式调度执行权并持续续租。
				if s.rdb != nil {
					lock := redissync.NewMutex(s.rdb, "studyroom:scheduler:leader", 30*time.Second)
					acquired, err := lock.TryLock(ctx)
					if err != nil {
						log.Printf("[scheduler] 获取调度锁失败，本轮跳过: %v", err)
						return
					}
					if !acquired {
						return
					}
					stopRenew := lock.StartAutoRenew(ctx, 10*time.Second, func(err error) {
						log.Printf("[scheduler] 调度锁续租失败，数据库条件更新继续兜底: %v", err)
					})
					defer func() { _ = lock.Unlock(context.Background()) }()
					defer stopRenew()
				}

				v, c, l := s.lifecycle.Tick(ctx)
				p := s.waitlist.PromoteTick(ctx)
				e := s.waitlist.ExpireTick(ctx)
				if v+c+l+p+e > 0 {
					log.Printf("[scheduler] 本轮处理: 违约=%d 完成=%d 离开超时=%d 候补递补=%d 候补过期=%d",
						v, c, l, p, e)
				}
			}()
		}
	}
}
