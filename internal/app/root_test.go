package app

import (
	"context"
	"testing"

	"github.com/go-to-k/delstack/internal/io"
)

func TestRootAction_Validation(t *testing.T) {
	io.NewLogger(false)

	tests := []struct {
		name    string
		action  *RootAction
		wantErr string
	}{
		{
			name:    "no stack names and not interactive mode",
			action:  NewRootAction(nil, "", "", false, false, true, 0),
			wantErr: "InvalidOptionError",
		},
		{
			name:    "stack names with interactive mode",
			action:  NewRootAction([]string{"Stack1"}, "", "", true, false, true, 0),
			wantErr: "InvalidOptionError",
		},
		{
			name:    "negative concurrency number",
			action:  NewRootAction([]string{"Stack1"}, "", "", false, false, true, -1),
			wantErr: "InvalidOptionError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Run(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
