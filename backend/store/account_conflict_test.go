package store

import (
	"testing"

	"windsurf-tools-wails/backend/models"
)

func TestAccountsConflict(t *testing.T) {
	t.Parallel()
	a := models.Account{Email: "A@Example.com", Token: "jwt-a"}
	b := models.Account{Email: "a@example.com", Token: "jwt-b"}
	if !AccountsConflict(a, b) {
		t.Fatal("same normalized email should conflict")
	}
	if AccountsConflict(models.Account{Email: "JWT #1"}, models.Account{Email: "JWT #2"}) {
		t.Fatal("placeholder emails should not match by email alone")
	}
	rt1 := models.Account{RefreshToken: " rt-1 "}
	rt2 := models.Account{RefreshToken: "rt-1"}
	if !AccountsConflict(rt1, rt2) {
		t.Fatal("same refresh token should conflict")
	}
	key1 := models.Account{WindsurfAPIKey: "k1"}
	key2 := models.Account{WindsurfAPIKey: "k1"}
	if !AccountsConflict(key1, key2) {
		t.Fatal("same api key should conflict")
	}
	tok := models.Account{Token: "same.jwt.here"}
	if !AccountsConflict(tok, tok) {
		t.Fatal("same jwt should conflict")
	}
	u1 := models.Account{Email: "user_Ab12cd34"}
	u2 := models.Account{Email: "USER_ab12cd34"}
	if !AccountsConflict(u1, u2) {
		t.Fatal("synthetic user_ id should conflict case-insensitively")
	}
}
