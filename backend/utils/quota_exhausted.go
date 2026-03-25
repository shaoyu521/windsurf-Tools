package utils

import (
	"strconv"
	"strings"

	"windsurf-tools-wails/backend/models"
)

// ParseQuotaPercentString 解析账号卡片上的日/周剩余字符串（如 "0.00%"）。
func ParseQuotaPercentString(s string) (v float64, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// WeeklyQuotaMissingBlocksUsage 对于带 weekly resetAt 的日/周额度套餐，
// 官方返回「daily 仍有值，但 weekly 缺失」时，实际请求常已被上游按周限耗尽拒绝。
// 这里不伪造展示字段，只在可用性判定上把这类账号视为不可用。
func WeeklyQuotaMissingBlocksUsage(acc *models.Account) bool {
	if acc == nil {
		return false
	}
	if strings.TrimSpace(acc.WeeklyRemaining) != "" {
		return false
	}
	if strings.TrimSpace(acc.WeeklyResetAt) == "" {
		return false
	}
	if daily, ok := ParseQuotaPercentString(acc.DailyRemaining); ok && daily <= 0.0001 {
		return false
	}
	return true
}

// AccountQuotaExhausted 根据已同步的额度字段判断是否「可用配额见底」。
// 规则：月/积分型 total>0 且 used>=total；或日、周剩余百分比任一≤0 即视为用尽；
// 对于带 weekly resetAt 的套餐，如果官方 weekly 字段缺失，也按不可用处理
// （服务端会在 weekly 耗尽时拒绝请求，即使 daily 仍有余量）。
func AccountQuotaExhausted(acc *models.Account) bool {
	if acc == nil {
		return false
	}
	if acc.TotalQuota > 0 && acc.UsedQuota >= acc.TotalQuota {
		return true
	}
	d, dOk := ParseQuotaPercentString(acc.DailyRemaining)
	w, wOk := ParseQuotaPercentString(acc.WeeklyRemaining)
	if dOk && d <= 0.0001 {
		return true
	}
	if wOk && w <= 0.0001 {
		return true
	}
	if WeeklyQuotaMissingBlocksUsage(acc) {
		return true
	}
	return false
}
