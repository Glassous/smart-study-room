package service

import (
	"context"
	"log"
	"time"
)

// Scheduler 后台周期调度器
// 每轮依次执行: 违约扫描 → 到时自动完成 → 临时离开超时处理
// (候补递补在 waitlist 模块并入后挂载到同一调度循环)
type Scheduler struct {
	lifecycle *LifecycleService
	interval  time.Duration
}

func NewScheduler(lifecycle *LifecycleService, interval time.Duration) *Scheduler {
	return &Scheduler{lifecycle: lifecycle, interval: interval}
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
			v, c, l := s.lifecycle.Tick(ctx)
			if v+c+l > 0 {
				log.Printf("[scheduler] 本轮处理: 违约=%d 完成=%d 离开超时=%d", v, c, l)
			}
		}
	}
}
