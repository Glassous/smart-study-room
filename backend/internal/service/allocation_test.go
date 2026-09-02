package service

import (
	"testing"

	"github.com/imicola/smart-study-room/backend/internal/model"
)

func b(v bool) *bool { return &v }

// 构造候选座位
func cand(seatNo, zone string, power, window bool, util float64) *model.CandidateSeat {
	return &model.CandidateSeat{
		Seat: model.Seat{
			ID: 1, SeatNo: seatNo, Zone: zone,
			HasPower: power, NearWindow: window, Status: model.SeatAvailable,
		},
		Util7d: util,
	}
}

// TC-A01: 偏好匹配得分(区域30+电源25+靠窗15)
func TestScoreCandidatesPreferenceMatch(t *testing.T) {
	req := &model.AllocationRequest{
		Zone: "quiet", HasPower: b(true), NearWindow: b(true),
	}
	candidates := []*model.CandidateSeat{
		cand("A01", "quiet", true, true, 0),   // 全匹配: 30+25+15 = 70
		cand("A02", "quiet", true, false, 0),  // 55
		cand("A03", "regular", false, false, 0), // 0
	}
	scored := ScoreCandidates(candidates, req)
	if scored[0].Seat.SeatNo != "A01" {
		t.Fatalf("最优应为 A01, 实际 %s", scored[0].Seat.SeatNo)
	}
	// A01 总分 = 70 + 0均衡 + 扰动(0~1) → 落在 [70,71)
	if scored[0].Score < 70 || scored[0].Score >= 71 {
		t.Fatalf("A01 评分应在[70,71), 实际 %.2f", scored[0].Score)
	}
	if scored[0].Reasons == nil || len(scored[0].Reasons) < 3 {
		t.Fatalf("A01 推荐理由应含 区域/电源/靠窗 三项: %v", scored[0].Reasons)
	}
}

// TC-A02: 热度均衡因子(同偏好下低利用率优先)
func TestScoreCandidatesUtilizationBalance(t *testing.T) {
	req := &model.AllocationRequest{Zone: "quiet"}
	candidates := []*model.CandidateSeat{
		cand("HOT", "quiet", false, false, 0.9),  // 30 - 18 = 12
		cand("COLD", "quiet", false, false, 0.1), // 30 - 2 = 28
	}
	scored := ScoreCandidates(candidates, req)
	if scored[0].Seat.SeatNo != "COLD" {
		t.Fatalf("低利用率座位应优先, 实际首位 %s", scored[0].Seat.SeatNo)
	}
	// 分差约 16, 扰动最多 1, 不可能翻转
	if scored[0].Score-scored[1].Score < 15 {
		t.Fatalf("分差异常: %.2f", scored[0].Score-scored[1].Score)
	}
}

// TC-A03: 排序为总分降序
func TestScoreCandidatesSortedDesc(t *testing.T) {
	req := &model.AllocationRequest{Zone: "quiet"}
	candidates := []*model.CandidateSeat{
		cand("M", "quiet", false, false, 0.5),
		cand("L", "quiet", false, false, 0.0),
		cand("H", "quiet", false, false, 1.0),
	}
	scored := ScoreCandidates(candidates, req)
	for i := 1; i < len(scored); i++ {
		if scored[i-1].Score < scored[i].Score {
			t.Fatalf("排序非降序: [%d]=%.2f < [%d]=%.2f", i-1, scored[i-1].Score, i, scored[i].Score)
		}
	}
}

// TC-A04: 无偏好时均衡因子仍然生效
func TestScoreCandidatesNoPreference(t *testing.T) {
	req := &model.AllocationRequest{}
	candidates := []*model.CandidateSeat{
		cand("HOT", "regular", false, false, 0.8),
		cand("COLD", "regular", false, false, 0.0),
	}
	scored := ScoreCandidates(candidates, req)
	if scored[0].Seat.SeatNo != "COLD" {
		t.Fatalf("无偏好时应选低利用率座位, 实际 %s", scored[0].Seat.SeatNo)
	}
	if len(scored[0].Reasons) == 0 {
		t.Fatalf("低利用率座位应有'分散选择'理由")
	}
}
