package utils

import (
	"testing"

	"windsurf-tools-wails/backend/models"
)

func TestAccountQuotaExhausted(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		acc  models.Account
		want bool
	}{
		{"nil guard", models.Account{}, false},
		{"monthly cap", models.Account{TotalQuota: 100, UsedQuota: 100}, true},
		{"monthly not full", models.Account{TotalQuota: 100, UsedQuota: 99}, false},
		{"daily zero only", models.Account{DailyRemaining: "0.00%"}, true},
		{"daily partial", models.Account{DailyRemaining: "12.00%"}, false},
		{"weekly zero only", models.Account{DailyRemaining: "100.00%", WeeklyRemaining: "0.00%"}, true},
		{"both zero", models.Account{DailyRemaining: "0%", WeeklyRemaining: "0.00%"}, true},
		{"daily zero weekly ok", models.Account{DailyRemaining: "0%", WeeklyRemaining: "50%"}, true},
		{"weekly missing with reset blocks usage", models.Account{DailyRemaining: "100.00%", WeeklyRemaining: "", WeeklyResetAt: "2026-03-29T08:00:00Z"}, true},
		{"weekly missing without reset stays unknown", models.Account{DailyRemaining: "100.00%", WeeklyRemaining: ""}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			acc := tc.acc
			if got := AccountQuotaExhausted(&acc); got != tc.want {
				t.Fatalf("AccountQuotaExhausted = %v, want %v", got, tc.want)
			}
		})
	}
}
