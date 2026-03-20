package utils

import "time"

// QuotaRefreshPolicy 与 windsurf-account-manager-simple 中 quotaRefreshPolicy 对齐，并扩展 Windows 常用策略
const (
	QuotaPolicyHybrid        = "hybrid"         // 满 24h 或美东换日
	QuotaPolicyInterval24h   = "interval_24h"   // 仅固定 24 小时
	QuotaPolicyUSCalendar    = "us_calendar"    // 仅美东日历跨日
	QuotaPolicyLocalCalendar = "local_calendar" // 本机时区日历跨日（Windows 区域时区）
	QuotaPolicyInterval1h    = "interval_1h"
	QuotaPolicyInterval6h    = "interval_6h"
	QuotaPolicyInterval12h   = "interval_12h"
	QuotaPolicyCustom        = "custom" // 配合 quota_custom_interval_minutes
)

func usDateNewYork(t time.Time) string {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return t.UTC().Format("2006-01-02")
	}
	return t.In(loc).Format("2006-01-02")
}

func localCalendarDate(t time.Time) string {
	return t.Local().Format("2006-01-02")
}

// ClampQuotaCustomIntervalMinutes 自定义同步间隔（分钟）：未设置或非法时用默认 360；最小 5，最大 7 天
func ClampQuotaCustomIntervalMinutes(m int) int {
	const def = 360
	if m <= 0 {
		return def
	}
	if m < 5 {
		return 5
	}
	if m > 10080 {
		return 10080
	}
	return m
}

// QuotaRefreshDue 判断是否需要拉取配额展示（lastQuotaUpdate 为 RFC3339；customIntervalMinutes 仅在 policy=custom 时生效）
func QuotaRefreshDue(lastQuotaUpdateISO, policy string, customIntervalMinutes int, now time.Time) bool {
	if lastQuotaUpdateISO == "" {
		return true
	}
	last, err := time.Parse(time.RFC3339, lastQuotaUpdateISO)
	if err != nil {
		return true
	}

	switch policy {
	case QuotaPolicyInterval24h:
		return now.Sub(last) >= 24*time.Hour
	case QuotaPolicyUSCalendar:
		return usDateNewYork(last) < usDateNewYork(now)
	case QuotaPolicyLocalCalendar:
		return localCalendarDate(last) < localCalendarDate(now)
	case QuotaPolicyInterval1h:
		return now.Sub(last) >= time.Hour
	case QuotaPolicyInterval6h:
		return now.Sub(last) >= 6*time.Hour
	case QuotaPolicyInterval12h:
		return now.Sub(last) >= 12*time.Hour
	case QuotaPolicyCustom:
		d := time.Duration(ClampQuotaCustomIntervalMinutes(customIntervalMinutes)) * time.Minute
		return now.Sub(last) >= d
	default:
		// hybrid：24 小时或美东新日历日
		if now.Sub(last) >= 24*time.Hour {
			return true
		}
		return usDateNewYork(last) < usDateNewYork(now)
	}
}
