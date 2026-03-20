package main

import (
	"testing"
	"windsurf-tools-wails/backend/models"
	"windsurf-tools-wails/backend/services"
)

func TestDerivePlanNameFromClaims(t *testing.T) {
	cases := []struct {
		name   string
		claims services.JWTClaims
		want   string
	}{
		{name: "pro flag", claims: services.JWTClaims{Pro: true}, want: "Pro"},
		{name: "teams pro", claims: services.JWTClaims{TeamsTier: "TEAMS_TIER_PRO"}, want: "Pro"},
		{name: "teams max", claims: services.JWTClaims{TeamsTier: "TEAMS_TIER_PRO_MAX"}, want: "Max"},
		{name: "trial tier", claims: services.JWTClaims{TeamsTier: "TEAMS_TIER_TRIAL"}, want: "Trial"},
		{name: "enterprise tier", claims: services.JWTClaims{TeamsTier: "TEAMS_TIER_ENTERPRISE"}, want: "Enterprise"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := derivePlanNameFromClaims(&tc.claims)
			if got != tc.want {
				t.Fatalf("derivePlanNameFromClaims() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestApplyAccountProfile(t *testing.T) {
	acc := &models.Account{
		Email:    "",
		Nickname: "",
		PlanName: "unknown",
	}
	daily := 78.0
	weekly := 88.0

	applyAccountProfile(acc, &services.AccountProfile{
		Email:                 "trial@example.com",
		Name:                  "Trial User",
		PlanName:              "Trial",
		TotalCredits:          53,
		UsedCredits:           14,
		DailyQuotaRemaining:   &daily,
		WeeklyQuotaRemaining:  &weekly,
		DailyResetAt:          "2026-03-21T08:00:00Z",
		WeeklyResetAt:         "2026-03-22T08:00:00Z",
		SubscriptionExpiresAt: "2026-03-29T08:00:00Z",
	})

	if acc.Email != "trial@example.com" {
		t.Fatalf("Email = %q", acc.Email)
	}
	if acc.Nickname != "Trial User" {
		t.Fatalf("Nickname = %q", acc.Nickname)
	}
	if acc.PlanName != "Trial" {
		t.Fatalf("PlanName = %q", acc.PlanName)
	}
	if acc.TotalQuota != 53 || acc.UsedQuota != 14 {
		t.Fatalf("quota = %d/%d", acc.UsedQuota, acc.TotalQuota)
	}
	if acc.DailyRemaining != "78.00%" || acc.WeeklyRemaining != "88.00%" {
		t.Fatalf("remaining = %q / %q", acc.DailyRemaining, acc.WeeklyRemaining)
	}
	if acc.DailyResetAt == "" || acc.WeeklyResetAt == "" {
		t.Fatal("reset timestamps should be set")
	}
	if acc.SubscriptionExpiresAt != "2026-03-29T08:00:00Z" {
		t.Fatalf("SubscriptionExpiresAt = %q", acc.SubscriptionExpiresAt)
	}
}
