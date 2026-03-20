package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"windsurf-tools-wails/backend/models"
	"windsurf-tools-wails/backend/services"
	"windsurf-tools-wails/backend/store"
	"windsurf-tools-wails/backend/utils"
)

type App struct {
	ctx                    context.Context
	store                  *store.Store
	windsurfSvc            *services.WindsurfService
	switchSvc              *services.SwitchService
	patchSvc               *services.PatchService
	cancelAutoRefresh      context.CancelFunc
	cancelAutoQuotaRefresh context.CancelFunc
	mu                     sync.Mutex
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	s, err := store.NewStore()
	if err != nil {
		fmt.Println("初始化存储失败:", err)
		return
	}
	a.store = s
	settings := a.store.GetSettings()
	proxyURL := ""
	if settings.ProxyEnabled && settings.ProxyURL != "" {
		proxyURL = settings.ProxyURL
	}
	a.windsurfSvc = services.NewWindsurfService(proxyURL)
	a.switchSvc = services.NewSwitchService()
	a.patchSvc = services.NewPatchService()
	if settings.AutoRefreshTokens {
		a.startAutoRefresh()
	}
	if settings.AutoRefreshQuotas {
		a.startAutoQuotaRefresh()
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancelAutoRefresh != nil {
		a.cancelAutoRefresh()
	}
	if a.cancelAutoQuotaRefresh != nil {
		a.cancelAutoQuotaRefresh()
	}
}

// ═══════════════════════════════════════
// 账号管理
// ═══════════════════════════════════════

func (a *App) GetAllAccounts() []models.Account {
	return a.store.GetAllAccounts()
}

func (a *App) DeleteAccount(id string) error {
	return a.store.DeleteAccount(id)
}

func (a *App) DeleteExpiredAccounts() (int, error) {
	accounts := a.store.GetAllAccounts()
	now := time.Now()
	deleted := 0
	for _, acc := range accounts {
		if acc.SubscriptionExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, acc.SubscriptionExpiresAt)
			if err == nil && t.Before(now) {
				if err := a.store.DeleteAccount(acc.ID); err == nil {
					deleted++
				}
			}
		}
	}
	return deleted, nil
}

// DeleteFreePlanAccounts 删除计划归类为 free 的账号（与 utils.PlanTone / 前端 getPlanTone 一致：名称含 free 或 basic）
func (a *App) DeleteFreePlanAccounts() (int, error) {
	accounts := a.store.GetAllAccounts()
	deleted := 0
	for _, acc := range accounts {
		if utils.PlanTone(acc.PlanName) != "free" {
			continue
		}
		if err := a.store.DeleteAccount(acc.ID); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// ═══════════════════════════════════════
// 批量导入 + 单个添加
// ═══════════════════════════════════════

type ImportResult struct {
	Email   string `json:"email"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type EmailPasswordItem struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	AltPassword string `json:"alt_password,omitempty"`
	Remark      string `json:"remark"`
}
type TokenItem struct {
	Token  string `json:"token"`
	Remark string `json:"remark"`
}
type APIKeyItem struct {
	APIKey string `json:"api_key"`
	Remark string `json:"remark"`
}
type JWTItem struct {
	JWT    string `json:"jwt"`
	Remark string `json:"remark"`
}

func (a *App) ImportByEmailPassword(items []EmailPasswordItem) []ImportResult {
	var results []ImportResult
	for _, item := range items {
		passwords := []string{item.Password}
		if item.AltPassword != "" && item.AltPassword != item.Password {
			passwords = append(passwords, item.AltPassword)
		}
		var resp *services.FirebaseSignInResp
		var err error
		var usedPassword string
		for _, pw := range passwords {
			if pw == "" {
				continue
			}
			resp, err = a.windsurfSvc.LoginWithEmail(item.Email, pw)
			if err == nil {
				usedPassword = pw
				break
			}
		}
		if err != nil {
			results = append(results, ImportResult{Email: item.Email, Success: false, Error: err.Error()})
			continue
		}
		nickname := item.Remark
		if nickname == "" {
			nickname = strings.Split(item.Email, "@")[0]
		}
		acc := models.NewAccount(item.Email, usedPassword, nickname)
		acc.Token = resp.IDToken
		acc.RefreshToken = resp.RefreshToken
		acc.TokenExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		acc.Remark = item.Remark
		a.enrichAccountInfoLite(acc)
		if err := a.store.AddAccount(*acc); err != nil {
			results = append(results, ImportResult{Email: item.Email, Success: false, Error: err.Error()})
			continue
		}
		results = append(results, ImportResult{Email: item.Email, Success: true})
	}
	return results
}

func (a *App) ImportByRefreshToken(items []TokenItem) []ImportResult {
	var results []ImportResult
	for i, item := range items {
		resp, err := a.windsurfSvc.RefreshToken(item.Token)
		if err != nil {
			results = append(results, ImportResult{
				Email: fmt.Sprintf("Token #%d", i+1), Success: false, Error: err.Error(),
			})
			continue
		}
		email, _ := a.windsurfSvc.GetAccountInfo(resp.IDToken)
		if email == "" {
			email = fmt.Sprintf("user_%s", resp.UserID[:minInt(8, len(resp.UserID))])
		}
		nickname := item.Remark
		if nickname == "" {
			nickname = strings.Split(email, "@")[0]
		}
		acc := models.NewAccount(email, "", nickname)
		acc.Token = resp.IDToken
		acc.RefreshToken = resp.RefreshToken
		acc.TokenExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		acc.Remark = item.Remark
		a.enrichAccountInfoLite(acc)
		if err := a.store.AddAccount(*acc); err != nil {
			results = append(results, ImportResult{Email: email, Success: false, Error: err.Error()})
			continue
		}
		results = append(results, ImportResult{Email: email, Success: true})
	}
	return results
}

func (a *App) ImportByAPIKey(items []APIKeyItem) []ImportResult {
	var results []ImportResult
	for i, item := range items {
		jwt, err := a.windsurfSvc.GetJWTByAPIKey(item.APIKey)
		if err != nil {
			results = append(results, ImportResult{
				Email: fmt.Sprintf("Key #%d", i+1), Success: false, Error: err.Error(),
			})
			continue
		}

		email := fmt.Sprintf("%s...%s", item.APIKey[:minInt(12, len(item.APIKey))],
			item.APIKey[maxInt(0, len(item.APIKey)-6):])

		acc := models.NewAccount(email, "", item.Remark)
		acc.Token = jwt
		acc.WindsurfAPIKey = item.APIKey
		acc.Remark = item.Remark
		a.enrichAccountInfoLite(acc)
		if item.Remark == "" {
			acc.Nickname = strings.Split(acc.Email, "@")[0]
		}

		if err := a.store.AddAccount(*acc); err != nil {
			results = append(results, ImportResult{Email: acc.Email, Success: false, Error: err.Error()})
			continue
		}
		results = append(results, ImportResult{Email: acc.Email, Success: true})
	}
	return results
}

func (a *App) ImportByJWT(items []JWTItem) []ImportResult {
	var results []ImportResult
	for i, item := range items {
		email := fmt.Sprintf("JWT #%d", i+1)
		acc := models.NewAccount(email, "", item.Remark)
		acc.Token = item.JWT
		acc.Remark = item.Remark
		a.enrichAccountInfoLite(acc)
		if item.Remark == "" {
			acc.Nickname = strings.Split(acc.Email, "@")[0]
		}

		if err := a.store.AddAccount(*acc); err != nil {
			results = append(results, ImportResult{Email: acc.Email, Success: false, Error: err.Error()})
			continue
		}
		results = append(results, ImportResult{Email: acc.Email, Success: true})
	}
	return results
}

// 单个添加
func (a *App) AddSingleAccount(mode string, value string, remark string) ImportResult {
	switch mode {
	case "api_key":
		items := []APIKeyItem{{APIKey: value, Remark: remark}}
		r := a.ImportByAPIKey(items)
		if len(r) > 0 {
			return r[0]
		}
	case "jwt":
		items := []JWTItem{{JWT: value, Remark: remark}}
		r := a.ImportByJWT(items)
		if len(r) > 0 {
			return r[0]
		}
	case "refresh_token":
		items := []TokenItem{{Token: value, Remark: remark}}
		r := a.ImportByRefreshToken(items)
		if len(r) > 0 {
			return r[0]
		}
	case "password":
		var cred struct {
			Email       string `json:"email"`
			Password    string `json:"password"`
			AltPassword string `json:"alt_password"`
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(value)), &cred); err != nil {
			return ImportResult{Email: "?", Success: false, Error: "邮箱密码格式错误"}
		}
		if cred.Email == "" || cred.Password == "" {
			return ImportResult{Email: "?", Success: false, Error: "请填写邮箱与密码"}
		}
		r := a.ImportByEmailPassword([]EmailPasswordItem{{
			Email: cred.Email, Password: cred.Password, AltPassword: cred.AltPassword, Remark: remark,
		}})
		if len(r) > 0 {
			return r[0]
		}
	}
	return ImportResult{Email: "?", Success: false, Error: "无效的导入类型"}
}

// ═══════════════════════════════════════
// 无感切号
// ═══════════════════════════════════════

func (a *App) SwitchAccount(id string) error {
	acc, err := a.store.GetAccount(id)
	if err != nil {
		return err
	}
	if acc.Token == "" {
		return fmt.Errorf("该账号没有可用的Token")
	}
	return a.switchSvc.SwitchAccount(acc.Token, acc.Email)
}

// AutoSwitchToNext 切到下一可用账号。planFilter：all 不限制；否则为 PlanTone 单值或逗号分隔多选（如 trial,pro）
func (a *App) AutoSwitchToNext(currentID string, planFilter string) (string, error) {
	accounts := a.store.GetAllAccounts()
	for _, acc := range accounts {
		if acc.ID == currentID {
			continue
		}
		if acc.Token == "" {
			continue
		}
		if acc.Status == "disabled" || acc.Status == "expired" {
			continue
		}
		if !utils.PlanFilterMatch(planFilter, acc.PlanName) {
			continue
		}
		err := a.switchSvc.SwitchAccount(acc.Token, acc.Email)
		if err == nil {
			return acc.Email, nil
		}
	}
	f := strings.TrimSpace(strings.ToLower(planFilter))
	if f != "" && f != "all" {
		return "", fmt.Errorf("在「%s」类型下没有可切换的账号", f)
	}
	return "", fmt.Errorf("没有可用的账号可以切换")
}

func (a *App) GetCurrentWindsurfAuth() (*services.WindsurfAuthJSON, error) {
	return a.switchSvc.GetCurrentAuth()
}

func (a *App) GetWindsurfAuthPath() (string, error) {
	return a.switchSvc.GetWindsurfAuthPath()
}

// ═══════════════════════════════════════
// 自动刷新 Token / JWT + 额度监控
// ═══════════════════════════════════════

func (a *App) startAutoRefresh() {
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelAutoRefresh = cancel
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.refreshAllTokens()
			}
		}
	}()
}

func (a *App) startAutoQuotaRefresh() {
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelAutoQuotaRefresh = cancel
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		a.refreshDueQuotas()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.refreshDueQuotas()
			}
		}
	}()
}

func (a *App) refreshDueQuotas() {
	a.mu.Lock()
	defer a.mu.Unlock()
	settings := a.store.GetSettings()
	if !settings.AutoRefreshQuotas {
		return
	}
	policy := strings.TrimSpace(settings.QuotaRefreshPolicy)
	if policy == "" {
		policy = utils.QuotaPolicyHybrid
	}
	now := time.Now()
	customMins := settings.QuotaCustomIntervalMinutes
	accounts := a.store.GetAllAccounts()
	for _, acc := range accounts {
		if !utils.QuotaRefreshDue(acc.LastQuotaUpdate, policy, customMins, now) {
			continue
		}
		if acc.WindsurfAPIKey == "" && acc.Token == "" && acc.RefreshToken == "" && (acc.Email == "" || acc.Password == "") {
			continue
		}
		copyAcc := acc
		a.syncAccountCredentials(&copyAcc)
		a.enrichAccountInfo(&copyAcc)
		copyAcc.LastQuotaUpdate = now.Format(time.RFC3339)
		_ = a.store.UpdateAccount(copyAcc)
	}
}

func (a *App) syncAccountCredentials(acc *models.Account) {
	if acc.WindsurfAPIKey != "" {
		if jwt, err := a.windsurfSvc.GetJWTByAPIKey(acc.WindsurfAPIKey); err == nil {
			acc.Token = jwt
		}
		return
	}
	if acc.RefreshToken != "" {
		if resp, err := a.windsurfSvc.RefreshToken(acc.RefreshToken); err == nil {
			acc.Token = resp.IDToken
			acc.RefreshToken = resp.RefreshToken
			acc.TokenExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
			return
		}
	}
	if acc.Email != "" && acc.Password != "" {
		if resp, err := a.windsurfSvc.LoginWithEmail(acc.Email, acc.Password); err == nil {
			acc.Token = resp.IDToken
			acc.RefreshToken = resp.RefreshToken
			acc.TokenExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		}
	}
}

func (a *App) RefreshAllTokens() map[string]string { return a.refreshAllTokens() }

func (a *App) refreshAllTokens() map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	results := make(map[string]string)
	accounts := a.store.GetAllAccounts()
	for _, acc := range accounts {
		if acc.WindsurfAPIKey != "" {
			jwt, err := a.windsurfSvc.GetJWTByAPIKey(acc.WindsurfAPIKey)
			if err != nil {
				results[acc.Email] = "JWT刷新失败: " + err.Error()
				continue
			}
			acc.Token = jwt
			a.enrichAccountInfo(&acc)
			acc.LastQuotaUpdate = time.Now().Format(time.RFC3339)
			_ = a.store.UpdateAccount(acc)
			results[acc.Email] = "JWT刷新成功"
			continue
		}
		if acc.RefreshToken != "" {
			resp, err := a.windsurfSvc.RefreshToken(acc.RefreshToken)
			if err != nil {
				results[acc.Email] = "Token刷新失败: " + err.Error()
				continue
			}
			acc.Token = resp.IDToken
			acc.RefreshToken = resp.RefreshToken
			acc.TokenExpiresAt = time.Now().Add(1 * time.Hour).Format(time.RFC3339)
			a.enrichAccountInfo(&acc)
			_ = a.store.UpdateAccount(acc)
			results[acc.Email] = "Token刷新成功"
			continue
		}
		results[acc.Email] = "无可用刷新凭证"
	}
	return results
}

// RefreshAccountQuota 手动同步单账号额度（同步凭证 + 拉取 profile，不校验策略间隔）
func (a *App) RefreshAccountQuota(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	acc, err := a.store.GetAccount(id)
	if err != nil {
		return err
	}
	if acc.WindsurfAPIKey == "" && acc.Token == "" && acc.RefreshToken == "" && (acc.Email == "" || acc.Password == "") {
		return fmt.Errorf("该账号没有可用于拉取额度的凭证")
	}
	copyAcc := *acc
	a.syncAccountCredentials(&copyAcc)
	a.enrichAccountInfo(&copyAcc)
	copyAcc.LastQuotaUpdate = time.Now().Format(time.RFC3339)
	return a.store.UpdateAccount(copyAcc)
}

// RefreshAllQuotas 手动同步全部账号额度（忽略 auto_refresh_quotas 与策略）
func (a *App) RefreshAllQuotas() map[string]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	results := make(map[string]string)
	now := time.Now().Format(time.RFC3339)
	for _, acc := range a.store.GetAllAccounts() {
		if acc.WindsurfAPIKey == "" && acc.Token == "" && acc.RefreshToken == "" && (acc.Email == "" || acc.Password == "") {
			results[labelAccountResult(acc)] = "跳过：无可用凭证"
			continue
		}
		copyAcc := acc
		a.syncAccountCredentials(&copyAcc)
		a.enrichAccountInfo(&copyAcc)
		copyAcc.LastQuotaUpdate = now
		if err := a.store.UpdateAccount(copyAcc); err != nil {
			results[labelAccountResult(acc)] = "失败: " + err.Error()
			continue
		}
		results[labelAccountResult(acc)] = "额度已同步"
	}
	return results
}

func labelAccountResult(acc models.Account) string {
	if acc.Email != "" {
		return acc.Email
	}
	return acc.ID
}

// ═══════════════════════════════════════
// 设置与代理
// ═══════════════════════════════════════

func (a *App) GetSettings() models.Settings { return a.store.GetSettings() }

func (a *App) UpdateSettings(settings models.Settings) error {
	if err := a.store.UpdateSettings(settings); err != nil {
		return err
	}
	proxyURL := ""
	if settings.ProxyEnabled && settings.ProxyURL != "" {
		proxyURL = settings.ProxyURL
	}
	a.windsurfSvc = services.NewWindsurfService(proxyURL)
	if settings.AutoRefreshTokens {
		if a.cancelAutoRefresh == nil {
			a.startAutoRefresh()
		}
	} else {
		if a.cancelAutoRefresh != nil {
			a.cancelAutoRefresh()
			a.cancelAutoRefresh = nil
		}
	}
	if settings.AutoRefreshQuotas {
		if a.cancelAutoQuotaRefresh == nil {
			a.startAutoQuotaRefresh()
		}
	} else {
		if a.cancelAutoQuotaRefresh != nil {
			a.cancelAutoQuotaRefresh()
			a.cancelAutoQuotaRefresh = nil
		}
	}
	return nil
}

// ═══════════════════════════════════════
// 辅助
// ═══════════════════════════════════════

// enrichAccountInfoLite 批量导入时使用：只做本地 JWT 解析，避免 RegisterUser / GetPlan / GetUserStatus 等串行请求拖死界面。
func (a *App) enrichAccountInfoLite(acc *models.Account) {
	if acc == nil {
		return
	}
	if acc.Token != "" {
		if claims, err := a.windsurfSvc.DecodeJWTClaims(acc.Token); err == nil {
			applyJWTClaims(acc, claims)
		}
	}
	if acc.Nickname == "" && acc.Email != "" {
		if at := strings.Index(acc.Email, "@"); at > 0 {
			acc.Nickname = acc.Email[:at]
		}
	}
	if acc.PlanName == "" {
		acc.PlanName = "unknown"
	}
}

func (a *App) enrichAccountInfo(acc *models.Account) {
	if acc == nil {
		return
	}

	if acc.Token != "" {
		if claims, err := a.windsurfSvc.DecodeJWTClaims(acc.Token); err == nil {
			applyJWTClaims(acc, claims)
		}
	}

	if acc.Token != "" && (acc.RefreshToken != "" || acc.Password != "") {
		if email, err := a.windsurfSvc.GetAccountInfo(acc.Token); err == nil && email != "" {
			acc.Email = email
		}
		if reg, err := a.windsurfSvc.RegisterUser(acc.Token); err == nil && reg != nil && reg.APIKey != "" {
			acc.WindsurfAPIKey = reg.APIKey
		}
		if plan, err := a.windsurfSvc.GetPlanStatusJSON(acc.Token); err == nil {
			applyAccountProfile(acc, plan)
		}
	}

	if acc.WindsurfAPIKey != "" {
		if profile, err := a.windsurfSvc.GetUserStatus(acc.WindsurfAPIKey); err == nil {
			applyAccountProfile(acc, profile)
		}
	}

	if acc.Nickname == "" && acc.Email != "" {
		acc.Nickname = strings.Split(acc.Email, "@")[0]
	}

	if acc.PlanName == "" {
		acc.PlanName = "unknown"
	}
}

func applyJWTClaims(acc *models.Account, claims *services.JWTClaims) {
	if claims == nil {
		return
	}
	if claims.Email != "" {
		acc.Email = claims.Email
	}
	if acc.Nickname == "" && claims.Name != "" {
		acc.Nickname = claims.Name
	}
	if plan := derivePlanNameFromClaims(claims); plan != "" && acc.PlanName == "unknown" {
		acc.PlanName = plan
	}
	if claims.TrialEnd != "" {
		acc.SubscriptionExpiresAt = claims.TrialEnd
	}
}

func applyAccountProfile(acc *models.Account, profile *services.AccountProfile) {
	if profile == nil {
		return
	}
	if profile.Email != "" {
		acc.Email = profile.Email
	}
	if profile.Name != "" && (acc.Nickname == "" || acc.Nickname == strings.Split(acc.Email, "@")[0]) {
		acc.Nickname = profile.Name
	}
	if profile.PlanName != "" {
		acc.PlanName = profile.PlanName
	}
	if profile.TotalCredits > 0 || profile.UsedCredits > 0 {
		acc.TotalQuota = profile.TotalCredits
		acc.UsedQuota = profile.UsedCredits
	}
	if profile.DailyQuotaRemaining != nil {
		acc.DailyRemaining = formatQuotaPercent(*profile.DailyQuotaRemaining)
	}
	if profile.WeeklyQuotaRemaining != nil {
		acc.WeeklyRemaining = formatQuotaPercent(*profile.WeeklyQuotaRemaining)
	}
	if profile.DailyResetAt != "" {
		acc.DailyResetAt = profile.DailyResetAt
	}
	if profile.WeeklyResetAt != "" {
		acc.WeeklyResetAt = profile.WeeklyResetAt
	}
	if profile.SubscriptionExpiresAt != "" {
		acc.SubscriptionExpiresAt = profile.SubscriptionExpiresAt
	}
}

func derivePlanNameFromClaims(claims *services.JWTClaims) string {
	if claims == nil {
		return ""
	}
	if claims.Pro {
		return "Pro"
	}
	teamsTier := strings.ToUpper(claims.TeamsTier)
	switch teamsTier {
	case "TEAMS_TIER_PRO":
		return "Pro"
	case "TEAMS_TIER_MAX", "TEAMS_TIER_PRO_MAX", "TEAMS_TIER_ULTIMATE":
		return "Max"
	case "TEAMS_TIER_ENTERPRISE":
		return "Enterprise"
	case "TEAMS_TIER_TRIAL":
		return "Trial"
	case "TEAMS_TIER_FREE":
		return "Free"
	}
	if strings.Contains(teamsTier, "TRIAL") {
		return "Trial"
	}
	if strings.Contains(teamsTier, "MAX") || strings.Contains(teamsTier, "ULTIMATE") {
		return "Max"
	}
	if strings.Contains(teamsTier, "ENTERPRISE") {
		return "Enterprise"
	}
	if strings.Contains(teamsTier, "PRO") {
		return "Pro"
	}
	if claims.TrialEnd != "" {
		if t, err := time.Parse(time.RFC3339Nano, claims.TrialEnd); err == nil {
			if t.After(time.Now()) {
				return "Trial"
			}
		}
	}
	return ""
}

func formatQuotaPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

// ═══════════════════════════════════════
// Patch
// ═══════════════════════════════════════

func (a *App) FindWindsurfPath() (string, error) { return a.patchSvc.FindWindsurfPath() }
func (a *App) ApplySeamlessPatch(p string) (*services.PatchResult, error) {
	return a.patchSvc.ApplyPatch(p)
}
func (a *App) RestoreSeamlessPatch(p string) error     { return a.patchSvc.RestorePatch(p) }
func (a *App) CheckPatchStatus(p string) (bool, error) { return a.patchSvc.CheckPatchStatus(p) }

// ═══════════════════════════════════════
// 工具函数
// ═══════════════════════════════════════

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
