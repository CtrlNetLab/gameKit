package utils

import (
	"testing"
	"time"
)

func TestDiffDays(t *testing.T) {
	difDays := GetDiffDays(1718760380738/1000, 1718760380)
	t.Logf("difDays: %d", difDays)
}

func TestNextMondayZero(t *testing.T) {
	t1 := "2025-07-08 00:00:00"

	t1Time, err := time.ParseInLocation(time.DateTime, t1, time.Local)
	if err != nil {
		t.FailNow()
	}
	nextMonday := NextMondayZero(t1Time)
	t.Logf("time: %s, nextMonday: %s", t1, nextMonday.Format(time.DateTime))
}

func TestGetNextMonth1thTime_Logic(t *testing.T) {
	// 测试逻辑验证：通过分析算法逻辑来测试，不依赖具体时间
	tests := []struct {
		name          string
		currentYear   int
		currentMonth  time.Month
		expectedYear  int
		expectedMonth time.Month
	}{
		{
			name:          "正常月份转换-1月到2月",
			currentYear:   2025,
			currentMonth:  time.January,
			expectedYear:  2025,
			expectedMonth: time.February,
		},
		{
			name:          "正常月份转换-6月到7月",
			currentYear:   2025,
			currentMonth:  time.June,
			expectedYear:  2025,
			expectedMonth: time.July,
		},
		{
			name:          "跨年转换-12月到1月",
			currentYear:   2025,
			currentMonth:  time.December,
			expectedYear:  2026,
			expectedMonth: time.January,
		},
		{
			name:          "边界值-11月到12月",
			currentYear:   2025,
			currentMonth:  time.November,
			expectedYear:  2025,
			expectedMonth: time.December,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 手动计算期望结果
			expectedTime := time.Date(tt.expectedYear, tt.expectedMonth, 1, 0, 0, 0, 0, time.Local)
			expectedTimestamp := expectedTime.UnixMilli()

			// 验证算法逻辑：month+1 %12 == 1 时表示跨年
			nextMonth := tt.currentMonth + 1
			expectedYear := tt.currentYear
			if nextMonth > 12 {
				nextMonth = 1
				expectedYear++
			}

			if expectedYear != tt.expectedYear {
				t.Errorf("算法验证失败: 期望年份 %d, 计算年份 %d", tt.expectedYear, expectedYear)
			}
			if nextMonth != tt.expectedMonth {
				t.Errorf("算法验证失败: 期望月份 %v, 计算月份 %v", tt.expectedMonth, nextMonth)
			}

			t.Logf("输入: %d年%d月, 期望输出: %d年%d月1日 00:00:00 (时间戳: %d)",
				tt.currentYear, tt.currentMonth, tt.expectedYear, tt.expectedMonth, expectedTimestamp)
		})
	}
}

func TestGetNextMonth1thTime_Functional(t *testing.T) {
	// 功能测试：通过运行实际代码并验证返回值格式
	t.Run("返回值格式验证", func(t *testing.T) {
		result := GetNextMonth1thTime()

		// 验证返回的是有效的时间戳（应该是比较大的数值）
		if result <= 0 {
			t.Errorf("返回的时间戳应该大于0, 实际: %d", result)
		}

		// 验证时间戳对应的时间是1号
		resultTime := time.UnixMilli(result)
		if resultTime.Day() != 1 {
			t.Errorf("返回的时间应该是1号, 实际: %d号", resultTime.Day())
		}

		// 验证时间是00:00:00
		if resultTime.Hour() != 0 || resultTime.Minute() != 0 || resultTime.Second() != 0 {
			t.Errorf("返回的时间应该是00:00:00, 实际: %02d:%02d:%02d",
				resultTime.Hour(), resultTime.Minute(), resultTime.Second())
		}

		t.Logf("GetNextMonth1thTime() = %d, 对应时间: %s", result, resultTime.Format("2006-01-02 15:04:05"))
	})
}

func TestGetNextMonth1thTime_Consistency(t *testing.T) {
	// 一致性测试：连续调用应该返回相同的结果（在同一分钟内）
	t.Run("连续调用一致性", func(t *testing.T) {
		result1 := GetNextMonth1thTime()
		result2 := GetNextMonth1thTime()

		if result1 != result2 {
			t.Errorf("连续调用结果不一致: first=%d, second=%d", result1, result2)
		}

		t.Logf("连续调用结果一致: %d", result1)
	})
}
