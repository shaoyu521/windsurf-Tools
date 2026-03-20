package store

import (
	"strings"
	"unicode"

	"windsurf-tools-wails/backend/models"
)

// AccountsConflict 判断两条账号记录是否视为同一身份（用于禁止重复导入）。
func AccountsConflict(existing, incoming models.Account) bool {
	if strings.TrimSpace(incoming.RefreshToken) != "" && strings.TrimSpace(existing.RefreshToken) != "" {
		if strings.TrimSpace(incoming.RefreshToken) == strings.TrimSpace(existing.RefreshToken) {
			return true
		}
	}
	if strings.TrimSpace(incoming.WindsurfAPIKey) != "" && strings.TrimSpace(existing.WindsurfAPIKey) != "" {
		if strings.TrimSpace(incoming.WindsurfAPIKey) == strings.TrimSpace(existing.WindsurfAPIKey) {
			return true
		}
	}
	if strings.TrimSpace(incoming.Token) != "" && strings.TrimSpace(existing.Token) != "" {
		if strings.TrimSpace(incoming.Token) == strings.TrimSpace(existing.Token) {
			return true
		}
	}
	if accountIdentityEmail(incoming.Email) && accountIdentityEmail(existing.Email) {
		if normalizeAccountEmail(incoming.Email) == normalizeAccountEmail(existing.Email) {
			return true
		}
	}
	return false
}

func normalizeAccountEmail(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// 仅对稳定身份邮箱做去重：真实邮箱、Refresh 回退的 user_xxx，排除 JWT#/Key#/Token# 等占位。
func accountIdentityEmail(email string) bool {
	e := strings.TrimSpace(email)
	if e == "" {
		return false
	}
	lower := strings.ToLower(e)
	if strings.HasPrefix(lower, "jwt #") || strings.HasPrefix(lower, "key #") || strings.HasPrefix(lower, "token #") {
		return false
	}
	if strings.ContainsRune(e, '@') {
		return true
	}
	if strings.HasPrefix(lower, "user_") {
		rest := lower[len("user_"):]
		if rest == "" {
			return false
		}
		for _, r := range rest {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
				return false
			}
		}
		return true
	}
	return false
}
