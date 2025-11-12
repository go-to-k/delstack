package app

import (
	"testing"

	"github.com/go-to-k/delstack/internal/io"
)

func TestNewStackDeleter(t *testing.T) {
	io.NewLogger(false)

	cases := []struct {
		name              string
		forceMode         bool
		concurrencyNumber int
		wantForceMode     bool
		wantConcurrency   int
	}{
		{
			name:              "create deleter with force mode",
			forceMode:         true,
			concurrencyNumber: 5,
			wantForceMode:     true,
			wantConcurrency:   5,
		},
		{
			name:              "create deleter without force mode",
			forceMode:         false,
			concurrencyNumber: UnspecifiedConcurrencyNumber,
			wantForceMode:     false,
			wantConcurrency:   UnspecifiedConcurrencyNumber,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			deleter := NewStackDeleter(tt.forceMode, tt.concurrencyNumber)

			if deleter.forceMode != tt.wantForceMode {
				t.Errorf("forceMode = %v, want %v", deleter.forceMode, tt.wantForceMode)
			}

			if deleter.concurrencyNumber != tt.wantConcurrency {
				t.Errorf("concurrencyNumber = %v, want %v", deleter.concurrencyNumber, tt.wantConcurrency)
			}
		})
	}
}
