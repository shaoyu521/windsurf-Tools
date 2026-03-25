package utils

import (
	"testing"
	"time"
	"windsurf-tools-wails/backend/models"
)

func TestQuotaRefreshDueEmptyLast(t *testing.T) {
	if !QuotaRefreshDue("", QuotaPolicyHybrid, 0, time.Now()) {
		t.Fatal("empty last should be due")
	}
}

func TestQuotaRefreshDueInterval24h(t *testing.T) {
	now := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	last := now.Add(-25 * time.Hour).Format(time.RFC3339)
	if !QuotaRefreshDue(last, QuotaPolicyInterval24h, 0, now) {
		t.Fatal("25h ago should be due")
	}
	last2 := now.Add(-12 * time.Hour).Format(time.RFC3339)
	if QuotaRefreshDue(last2, QuotaPolicyInterval24h, 0, now) {
		t.Fatal("12h ago should not be due")
	}
}

func TestQuotaRefreshDueLocalCalendar(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 3, 22, 1, 0, 0, 0, loc)
	last := time.Date(2026, 3, 21, 23, 0, 0, 0, loc).Format(time.RFC3339)
	if !QuotaRefreshDue(last, QuotaPolicyLocalCalendar, 0, now) {
		t.Fatal("local calendar day change should be due")
	}
	sameDay := time.Date(2026, 3, 22, 2, 0, 0, 0, loc).Format(time.RFC3339)
	if QuotaRefreshDue(sameDay, QuotaPolicyLocalCalendar, 0, now) {
		t.Fatal("same local day should not be due")
	}
}

func TestQuotaRefreshDueInterval1h(t *testing.T) {
	now := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	last := now.Add(-61 * time.Minute).Format(time.RFC3339)
	if !QuotaRefreshDue(last, QuotaPolicyInterval1h, 0, now) {
		t.Fatal("61m ago should be due for 1h policy")
	}
}

func TestQuotaRefreshDueCustom(t *testing.T) {
	now := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	last := now.Add(-31 * time.Minute).Format(time.RFC3339)
	if !QuotaRefreshDue(last, QuotaPolicyCustom, 30, now) {
		t.Fatal("31m ago should be due for 30min custom")
	}
	if QuotaRefreshDue(now.Add(-20*time.Minute).Format(time.RFC3339), QuotaPolicyCustom, 30, now) {
		t.Fatal("20m ago should not be due for 30min custom")
	}
}

func TestClampQuotaCustomIntervalMinutes(t *testing.T) {
	if ClampQuotaCustomIntervalMinutes(0) != 360 {
		t.Fatal("zero should default to 360")
	}
	if ClampQuotaCustomIntervalMinutes(3) != 5 {
		t.Fatal("below 5 should clamp to 5")
	}
	if ClampQuotaCustomIntervalMinutes(20000) != 10080 {
		t.Fatal("above 7d should clamp")
	}
}

func TestQuotaRefreshDueAfterOfficialReset_DailyReached(t *testing.T) {
	now := time.Date(2026, 3, 25, 8, 5, 0, 0, time.UTC)
	acc := models.Account{
		DailyRemaining:  "0.00%",
		DailyResetAt:    "2026-03-25T08:00:00Z",
		LastQuotaUpdate: "2026-03-25T07:30:00Z",
	}
	if !QuotaRefreshDueAfterOfficialReset(acc, now) {
		t.Fatal("expected official daily reset to force refresh")
	}
}

func TestQuotaRefreshDueAfterOfficialReset_WeeklyMissingReachedForces(t *testing.T) {
	now := time.Date(2026, 3, 25, 8, 5, 0, 0, time.UTC)
	acc := models.Account{
		DailyRemaining:  "65.00%",
		WeeklyRemaining: "",
		WeeklyResetAt:   "2026-03-25T08:00:00Z",
		LastQuotaUpdate: "2026-03-25T07:30:00Z",
	}
	if !QuotaRefreshDueAfterOfficialReset(acc, now) {
		t.Fatal("missing weekly quota after official reset should force refresh")
	}
}

func TestQuotaRefreshDueAfterOfficialReset_WeeklyMissingBeforeResetDoesNotForce(t *testing.T) {
	now := time.Date(2026, 3, 25, 7, 55, 0, 0, time.UTC)
	acc := models.Account{
		DailyRemaining:  "65.00%",
		WeeklyRemaining: "",
		WeeklyResetAt:   "2026-03-25T08:00:00Z",
		LastQuotaUpdate: "2026-03-25T07:30:00Z",
	}
	if QuotaRefreshDueAfterOfficialReset(acc, now) {
		t.Fatal("missing weekly quota before official reset should not force refresh")
	}
}

func TestNextQuotaResetWakeDelayForExhausted(t *testing.T) {
	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC)
	base := 30 * time.Second
	acc := models.Account{
		DailyRemaining: "0.00%",
		DailyResetAt:   "2026-03-25T08:00:10Z",
	}
	got := NextQuotaResetWakeDelayForExhausted(acc, now, base)
	if got != 10*time.Second {
		t.Fatalf("NextQuotaResetWakeDelayForExhausted = %v, want 10s", got)
	}
}

func TestNextQuotaResetWakeDelayForMissingWeeklyBlockedUsage(t *testing.T) {
	now := time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC)
	base := 30 * time.Second
	acc := models.Account{
		DailyRemaining:  "100.00%",
		WeeklyRemaining: "",
		WeeklyResetAt:   "2026-03-25T08:00:12Z",
	}
	got := NextQuotaResetWakeDelayForExhausted(acc, now, base)
	if got != 12*time.Second {
		t.Fatalf("NextQuotaResetWakeDelayForExhausted = %v, want 12s", got)
	}
}
