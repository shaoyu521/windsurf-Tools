package utils

import (
	"testing"
	"time"
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
