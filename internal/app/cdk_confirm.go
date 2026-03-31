package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
)

type CdkStackConfirmer struct {
	forceMode bool
}

func NewCdkStackConfirmer(forceMode bool) *CdkStackConfirmer {
	return &CdkStackConfirmer{forceMode: forceMode}
}

// ConfirmTPStacks checks for TerminationProtection stacks and handles confirmation.
// Returns an error if TP stacks are found without forceMode.
// Returns false if the user cancels the TP confirmation prompt.
func (c *CdkStackConfirmer) ConfirmTPStacks(stacks []cdk.StackInfo) (bool, error) {
	tpStacks := filterTPStacks(stacks)
	if len(tpStacks) == 0 {
		return true, nil
	}

	if !c.forceMode {
		return false, fmt.Errorf("TerminationProtectionError: %s", joinStackNames(tpStacks))
	}

	return c.showTPConfirmation(tpStacks), nil
}

// ConfirmDeletion shows the final deletion confirmation prompt.
func (c *CdkStackConfirmer) ConfirmDeletion(stacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks will be deleted:\n")
	for _, s := range stacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintln(os.Stderr)

	return io.GetYesNo("Are you sure you want to delete these stacks?")
}

func (c *CdkStackConfirmer) showTPConfirmation(tpStacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks have TerminationProtection enabled:\n")
	for _, s := range tpStacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintf(os.Stderr, "\nTerminationProtection will be disabled before deletion.\n")
	return io.GetYesNo("Do you want to proceed?")
}

func filterTPStacks(stacks []cdk.StackInfo) []cdk.StackInfo {
	var tp []cdk.StackInfo
	for _, s := range stacks {
		if s.TerminationProtection {
			tp = append(tp, s)
		}
	}
	return tp
}

func joinStackNames(stacks []cdk.StackInfo) string {
	names := make([]string, len(stacks))
	for i, s := range stacks {
		names[i] = s.StackName
	}
	return strings.Join(names, ", ")
}
