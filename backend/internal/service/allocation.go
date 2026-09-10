package service

import (
	"context"
	"errors"
	"math/rand"
	"sort"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// 评分权重(与需求/设计文档一致)
const (
	ScoreZoneMatch   = 30.0 // 区域偏好匹配
	ScorePowerMatch  = 25.0 // 电源偏好匹配
	ScoreWindowMatch = 15.0 // 靠窗偏好匹配
	ScoreUtilFactor  = 20.0 // 热度均衡因子系数
)

var ErrNoCandidate = errors.New("当前时段没有满足条件的空座位，可尝试加入候补")

// AllocationService 智能自动分配服务
type AllocationService struct {
	seats  *repository.SeatRepo
	rooms  *repository.RoomRepo
	users  *repository.UserRepo
	res    *repository.ReservationRepo
	resSvc *ReservationService
}

func NewAllocationService(
	seats *repository.SeatRepo,
	rooms *repository.RoomRepo,
	users *repository.UserRepo,
	res *repository.ReservationRepo,
	resSvc *ReservationService,
) *AllocationService {
	return &AllocationService{seats: seats, rooms: rooms, users: users, res: res, resSvc: resSvc}
}

// ScoreCandidates 纯函数评分: 偏好匹配分 + 热度均衡因子 + 微小扰动
// 独立成纯函数便于单元测试
func ScoreCandidates(candidates []*model.CandidateSeat, req *model.AllocationRequest) []*model.ScoredSeat {
	result := make([]*model.ScoredSeat, 0, len(candidates))
	for _, c := range candidates {
		var score float64
		var reasons []string

		// 偏好匹配(未指定偏好时不加分, 交给均衡因子)
		if req.Zone != "" {
			if c.Zone == req.Zone {
				score += ScoreZoneMatch
				reasons = append(reasons, "匹配区域偏好("+zoneName(c.Zone)+")")
			}
		}
		if req.HasPower != nil && *req.HasPower {
			if c.HasPower {
				score += ScorePowerMatch
				reasons = append(reasons, "带电源插座")
			}
		}
		if req.NearWindow != nil && *req.NearWindow {
			if c.NearWindow {
				score += ScoreWindowMatch
				reasons = append(reasons, "靠窗采光好")
			}
		}

		// 热度均衡: 利用率越高扣分越多(削峰填谷, 分散热门座位压力)
		score -= c.Util7d * ScoreUtilFactor
		if c.Util7d <= 0.2 {
			reasons = append(reasons, "近期使用率低, 分散选择")
		}

		// 微小随机扰动(0~1): 打散同分座位, 避免全涌向同一排
		score += rand.Float64()

		result = append(result, &model.ScoredSeat{
			Seat:    &c.Seat,
			Score:   score,
			Reasons: reasons,
		})
	}
	// 总分降序
	sort.SliceStable(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	return result
}

func zoneName(z string) string {
	switch z {
	case model.ZoneQuiet:
		return "静音区"
	case model.ZoneDiscussion:
		return "研讨区"
	case model.ZoneComputer:
		return "机房区"
	default:
		return "普通区"
	}
}

// AutoAllocate 推荐或直接分配
func (s *AllocationService) AutoAllocate(ctx context.Context, userID int64, req *model.AllocationRequest) (*model.AllocationResponse, error) {
	// 复用预约核心的时段/信用/房间校验
	if _, _, err := validateSlot(req.Date, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	resp := &model.AllocationResponse{}
	topN := req.TopN
	if topN <= 0 {
		topN = 5
	}

	candidates, err := s.seats.ListCandidates(ctx, req.RoomID, req.Date,
		padTime(req.StartTime), padTime(req.EndTime))
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, ErrNoCandidate
	}

	// 同用户该时段已有预约则无法分配(任一座位检测结果一致)
	if len(candidates) > 0 {
		if _, userConflict, err := s.res.HasConflict(ctx, candidates[0].ID, userID, req.Date,
			padTime(req.StartTime), padTime(req.EndTime)); err == nil && userConflict {
			return nil, ErrUserConflict
		}
	}

	scored := ScoreCandidates(candidates, req)
	if len(scored) > topN {
		scored = scored[:topN]
	}
	resp.Recommendations = scored

	// 确认分配: 取最高分座位创建预约(source=auto)
	if req.AutoBook && len(scored) > 0 {
		best := scored[0].Seat
		view, err := s.resSvc.CreateWithSource(ctx, userID, &model.CreateReservationRequest{
			SeatID:    best.ID,
			Date:      req.Date,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
		}, model.SrcAuto)
		if err != nil {
			return nil, err
		}
		resp.Booked = true
		resp.Reservation = view
	}
	return resp, nil
}
