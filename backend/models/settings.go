package models

// Settings 全局设置
type Settings struct {
	ProxyEnabled               bool   `json:"proxy_enabled"`
	ProxyURL                   string `json:"proxy_url"`
	WindsurfPath               string `json:"windsurf_path"`
	ConcurrentLimit            int    `json:"concurrent_limit"`
	SeamlessSwitch             bool   `json:"seamless_switch"`
	AutoRefreshTokens          bool   `json:"auto_refresh_tokens"`
	AutoRefreshQuotas          bool   `json:"auto_refresh_quotas"`
	QuotaRefreshPolicy         string `json:"quota_refresh_policy"`          // hybrid | interval_* | us_calendar | local_calendar | custom
	QuotaCustomIntervalMinutes int    `json:"quota_custom_interval_minutes"` // 仅 policy=custom 时使用，默认由后端钳制
	// AutoSwitchPlanFilter 无感「下一席位」计划池：all 不限制；否则逗号分隔多选，如 trial,pro（与 PlanTone 一致）
	AutoSwitchPlanFilter string `json:"auto_switch_plan_filter"`
}

func DefaultSettings() Settings {
	return Settings{
		ProxyEnabled:               false,
		ConcurrentLimit:            5,
		SeamlessSwitch:             false,
		AutoRefreshTokens:          false,
		AutoRefreshQuotas:          false,
		QuotaRefreshPolicy:         "hybrid",
		QuotaCustomIntervalMinutes: 360,
		AutoSwitchPlanFilter:       "all",
	}
}
