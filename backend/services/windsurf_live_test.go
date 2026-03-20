package services

import (
	"os"
	"testing"
)

func TestLiveAPIKeyStatus(t *testing.T) {
	apiKey := os.Getenv("WS_LIVE_API_KEY")
	if apiKey == "" {
		t.Skip("WS_LIVE_API_KEY is not set")
	}

	svc := NewWindsurfService("")
	jwt, err := svc.GetJWTByAPIKey(apiKey)
	if err != nil {
		t.Fatalf("GetJWTByAPIKey() error = %v", err)
	}
	if jwt == "" {
		t.Fatal("GetJWTByAPIKey() returned empty JWT")
	}

	claims, err := svc.DecodeJWTClaims(jwt)
	if err != nil {
		t.Fatalf("DecodeJWTClaims() error = %v", err)
	}
	if claims.Email == "" {
		t.Fatal("DecodeJWTClaims() returned empty email")
	}

	profile, err := svc.GetUserStatus(apiKey)
	if err != nil {
		t.Fatalf("GetUserStatus() error = %v", err)
	}
	if profile == nil {
		t.Fatal("GetUserStatus() returned nil profile")
	}
	if profile.Email == "" {
		t.Fatal("GetUserStatus() returned empty email")
	}
	if profile.PlanName == "" {
		t.Fatal("GetUserStatus() returned empty plan name")
	}
	if profile.DailyQuotaRemaining == nil || profile.WeeklyQuotaRemaining == nil {
		t.Fatalf("quota remaining missing: daily=%v weekly=%v", profile.DailyQuotaRemaining, profile.WeeklyQuotaRemaining)
	}
}
