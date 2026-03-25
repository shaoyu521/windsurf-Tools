package utils

import (
	"strings"
	"time"
	"windsurf-tools-wails/backend/models"
)

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

func quotaFieldResetReached(remaining, resetAt, lastQuotaUpdateISO string, now time.Time) bool {
	value, ok := ParseQuotaPercentString(remaining)
	if !ok || value > 0.0001 {
		return false
	}
	resetAt = strings.TrimSpace(resetAt)
	if resetAt == "" {
		return false
	}
	resetTime, err := time.Parse(time.RFC3339, resetAt)
	if err != nil || now.Before(resetTime) {
		return false
	}
	if strings.TrimSpace(lastQuotaUpdateISO) == "" {
		return true
	}
	lastUpdate, err := time.Parse(time.RFC3339, lastQuotaUpdateISO)
	if err != nil {
		return true
	}
	return lastUpdate.Before(resetTime)
}

func weeklyMissingResetReached(acc models.Account, now time.Time) bool {
	if !WeeklyQuotaMissingBlocksUsage(&acc) {
		return false
	}
	resetAt := strings.TrimSpace(acc.WeeklyResetAt)
	if resetAt == "" {
		return false
	}
	resetTime, err := time.Parse(time.RFC3339, resetAt)
	if err != nil || now.Before(resetTime) {
		return false
	}
	if strings.TrimSpace(acc.LastQuotaUpdate) == "" {
		return true
	}
	lastUpdate, err := time.Parse(time.RFC3339, acc.LastQuotaUpdate)
	if err != nil {
		return true
	}
	return lastUpdate.Before(resetTime)
}

// QuotaRefreshDueAfterOfficialReset 当官方返回的日/周 resetAt 已到且本地快照仍显示该额度为 0 时，
// 强制触发一次官方刷新，避免继续沿用旧快照。
func QuotaRefreshDueAfterOfficialReset(acc models.Account, now time.Time) bool {
	return quotaFieldResetReached(acc.DailyRemaining, acc.DailyResetAt, acc.LastQuotaUpdate, now) ||
		quotaFieldResetReached(acc.WeeklyRemaining, acc.WeeklyResetAt, acc.LastQuotaUpdate, now) ||
		weeklyMissingResetReached(acc, now)
}

func quotaFieldResetWakeDelay(remaining, resetAt string, now time.Time, base time.Duration) (time.Duration, bool) {
	value, ok := ParseQuotaPercentString(remaining)
	if !ok || value > 0.0001 {
		return base, false
	}
	resetAt = strings.TrimSpace(resetAt)
	if resetAt == "" {
		return base, false
	}
	resetTime, err := time.Parse(time.RFC3339, resetAt)
	if err != nil {
		return base, false
	}
	delay := resetTime.Sub(now)
	if delay <= 0 {
		return 0, true
	}
	if delay < base {
		return delay, true
	}
	return base, false
}

// NextQuotaResetWakeDelayForExhausted 当当前账号显示日/周额度为 0 且官方 resetAt 即将到来时，
// 让热轮询在 resetAt 附近提前醒来，而不是固定等完整的热轮询间隔。
func NextQuotaResetWakeDelayForExhausted(acc models.Account, now time.Time, base time.Duration) time.Duration {
	best := base
	if delay, ok := quotaFieldResetWakeDelay(acc.DailyRemaining, acc.DailyResetAt, now, best); ok && delay < best {
		best = delay
	}
	if delay, ok := quotaFieldResetWakeDelay(acc.WeeklyRemaining, acc.WeeklyResetAt, now, best); ok && delay < best {
		best = delay
	}
	if WeeklyQuotaMissingBlocksUsage(&acc) {
		if delay, ok := quotaFieldResetWakeDelay("0%", acc.WeeklyResetAt, now, best); ok && delay < best {
			best = delay
		}
	}
	if best < 0 {
		return 0
	}
	return best
}
