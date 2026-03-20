package utils

import "strings"

// PlanTone 与前端 src/utils/account.ts getPlanTone 对齐，用于号池分类与无感切号筛选
func PlanTone(planName string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(planName), "_", " "))
	if normalized == "" || normalized == "unknown" {
		return "unknown"
	}
	if strings.Contains(normalized, "trial") {
		return "trial"
	}
	if strings.Contains(normalized, "max") || strings.Contains(normalized, "ultimate") {
		return "max"
	}
	if strings.Contains(normalized, "enterprise") {
		return "enterprise"
	}
	if strings.Contains(normalized, "team") {
		return "team"
	}
	if strings.Contains(normalized, "pro") {
		return "pro"
	}
	if strings.Contains(normalized, "free") || strings.Contains(normalized, "basic") {
		return "free"
	}
	return "unknown"
}

// PlanFilterMatch filter 为 all / 空 时不过滤；单值 pro 等与 PlanTone 相等即匹配；
// 逗号分隔时为多选（如 trial,pro），账号计划属于任一即匹配。
func PlanFilterMatch(filter, planName string) bool {
	f := strings.TrimSpace(strings.ToLower(filter))
	if f == "" || f == "all" {
		return true
	}
	f = strings.ReplaceAll(f, "，", ",")
	tone := PlanTone(planName)
	for _, part := range strings.Split(f, ",") {
		p := strings.TrimSpace(strings.ToLower(part))
		if p == "" || p == "all" {
			return true
		}
		if p == tone {
			return true
		}
	}
	return false
}
