package io

import (
	"testing"
)

func TestGetYesNo_AutoYes(t *testing.T) {
	t.Cleanup(func() {
		AutoYes = false
	})

	AutoYes = true

	got := GetYesNo("Do you want to proceed?")
	if !got {
		t.Errorf("GetYesNo with AutoYes=true should return true, got false")
	}
}
