package app

import (
	"testing"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
)

func TestCdkStackConfirmer_ConfirmTPStacks(t *testing.T) {
	io.NewLogger(false)

	tpStack := cdk.StackInfo{StackName: "TPStack", Region: "us-east-1", TerminationProtection: true}
	normalStack := cdk.StackInfo{StackName: "NormalStack", Region: "us-east-1", TerminationProtection: false}

	t.Run("no TP stacks returns true", func(t *testing.T) {
		confirmer := NewCdkStackConfirmer(false)
		ok, err := confirmer.ConfirmTPStacks([]cdk.StackInfo{normalStack})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("empty stacks returns true", func(t *testing.T) {
		confirmer := NewCdkStackConfirmer(false)
		ok, err := confirmer.ConfirmTPStacks([]cdk.StackInfo{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("TP stacks without forceMode returns error", func(t *testing.T) {
		confirmer := NewCdkStackConfirmer(false)
		_, err := confirmer.ConfirmTPStacks([]cdk.StackInfo{tpStack, normalStack})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); got != "TerminationProtectionError: TPStack" {
			t.Errorf("error = %q, want %q", got, "TerminationProtectionError: TPStack")
		}
	})

	t.Run("multiple TP stacks without forceMode lists all names", func(t *testing.T) {
		tp2 := cdk.StackInfo{StackName: "TPStack2", Region: "ap-northeast-1", TerminationProtection: true}
		confirmer := NewCdkStackConfirmer(false)
		_, err := confirmer.ConfirmTPStacks([]cdk.StackInfo{tpStack, tp2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got := err.Error(); got != "TerminationProtectionError: TPStack, TPStack2" {
			t.Errorf("error = %q, want %q", got, "TerminationProtectionError: TPStack, TPStack2")
		}
	})

	t.Run("TP stacks with forceMode and AutoYes returns true", func(t *testing.T) {
		io.AutoYes = true
		defer func() { io.AutoYes = false }()

		confirmer := NewCdkStackConfirmer(true)
		ok, err := confirmer.ConfirmTPStacks([]cdk.StackInfo{tpStack, normalStack})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})
}

func TestCdkStackConfirmer_ConfirmDeletion(t *testing.T) {
	io.NewLogger(false)

	t.Run("returns true with AutoYes", func(t *testing.T) {
		io.AutoYes = true
		defer func() { io.AutoYes = false }()

		confirmer := NewCdkStackConfirmer(false)
		stacks := []cdk.StackInfo{
			{StackName: "StackA", Region: "us-east-1"},
			{StackName: "StackB", Region: "ap-northeast-1"},
		}
		if !confirmer.ConfirmDeletion(stacks) {
			t.Error("expected true, got false")
		}
	})
}

func TestCdkStackConfirmer_filterTPStacks(t *testing.T) {
	confirmer := NewCdkStackConfirmer(false)
	stacks := []cdk.StackInfo{
		{StackName: "TPStack", Region: "us-east-1", TerminationProtection: true},
		{StackName: "NormalStack", Region: "us-east-1", TerminationProtection: false},
		{StackName: "TPStack2", Region: "ap-northeast-1", TerminationProtection: true},
	}

	tp := confirmer.filterTPStacks(stacks)
	if len(tp) != 2 {
		t.Fatalf("expected 2 TP stacks, got %d", len(tp))
	}
	if tp[0].StackName != "TPStack" || tp[1].StackName != "TPStack2" {
		t.Errorf("unexpected TP stacks: %v", tp)
	}
}

func TestCdkStackConfirmer_joinStackNames(t *testing.T) {
	confirmer := NewCdkStackConfirmer(false)

	t.Run("single stack", func(t *testing.T) {
		stacks := []cdk.StackInfo{{StackName: "StackA"}}
		if got := confirmer.joinStackNames(stacks); got != "StackA" {
			t.Errorf("got %q, want %q", got, "StackA")
		}
	})

	t.Run("multiple stacks", func(t *testing.T) {
		stacks := []cdk.StackInfo{{StackName: "StackA"}, {StackName: "StackB"}, {StackName: "StackC"}}
		if got := confirmer.joinStackNames(stacks); got != "StackA, StackB, StackC" {
			t.Errorf("got %q, want %q", got, "StackA, StackB, StackC")
		}
	})

	t.Run("empty stacks", func(t *testing.T) {
		if got := confirmer.joinStackNames([]cdk.StackInfo{}); got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})
}
