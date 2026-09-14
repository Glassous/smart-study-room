package service

import (
	"testing"
	"time"
)

// TC-V01~V06: 时段校验 validateSlot 表驱动测试
func TestValidateSlot(t *testing.T) {
	cases := []struct {
		name    string
		date    string
		start   string
		end     string
		wantErr bool
	}{
		{"正常时段", "2026-09-10", "09:00", "12:00", false},
		{"半小时粒度", "2026-09-10", "09:30", "11:00", false},
		{"整八小时上限", "2026-09-10", "08:00", "16:00", false},
		{"超过八小时", "2026-09-10", "08:00", "16:30", true},
		{"粒度非法15分钟", "2026-09-10", "09:15", "10:45", true},
		{"开始等于结束", "2026-09-10", "09:00", "09:00", true},
		{"开始晚于结束", "2026-09-10", "14:00", "09:00", true},
		{"日期格式错误", "20260910", "09:00", "10:00", true},
		{"时间格式错误", "2026-09-10", "9:00", "10:00", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st, en, err := validateSlot(tc.date, tc.start, tc.end)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateSlot(%s,%s,%s) error=%v, wantErr=%v",
					tc.date, tc.start, tc.end, err, tc.wantErr)
			}
			if !tc.wantErr {
				d := en.Sub(st)
				if d <= 0 || d > 8*time.Hour+time.Minute {
					t.Fatalf("时长异常: %v", d)
				}
			}
		})
	}
}

// TC-V07: padTime 秒位补齐
func TestPadTime(t *testing.T) {
	cases := []struct{ in, want string }{
		{"09:00", "09:00:00"},
		{"09:00:30", "09:00:30"},
		{"23:59", "23:59:00"},
	}
	for _, tc := range cases {
		if got := padTime(tc.in); got != tc.want {
			t.Errorf("padTime(%s)=%s, want %s", tc.in, got, tc.want)
		}
	}
}
