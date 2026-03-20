package utils

import "testing"

func TestPlanTone(t *testing.T) {
	if PlanTone("Pro") != "pro" {
		t.Fatal()
	}
	if PlanTone("teams_ultimate") != "max" {
		t.Fatal()
	}
	if PlanFilterMatch("all", "anything") != true {
		t.Fatal()
	}
	if PlanFilterMatch("pro", "Pro") != true {
		t.Fatal()
	}
	if PlanFilterMatch("pro", "Teams") != false {
		t.Fatal()
	}
	if !PlanFilterMatch("trial,pro", "Pro") || !PlanFilterMatch("trial,pro", "Trial") {
		t.Fatal("multi filter should match trial or pro")
	}
	if PlanFilterMatch("trial,pro", "Teams") {
		t.Fatal("team should not match trial,pro")
	}
	if PlanTone("Free") != "free" || PlanTone("basic plan") != "free" {
		t.Fatal("free/basic tone")
	}
}
