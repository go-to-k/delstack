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

// Confirm runs the full confirmation flow: TP confirmation (if needed) then deletion confirmation.
// Returns an error if TP stacks are found without forceMode.
// Returns false if the user cancels any confirmation prompt.
func (c *CdkStackConfirmer) Confirm(stacks []cdk.StackInfo) (bool, error) {
	ok, err := c.confirmTPStacks(stacks)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	return c.confirmDeletion(stacks), nil
}

func (c *CdkStackConfirmer) confirmTPStacks(stacks []cdk.StackInfo) (bool, error) {
	tpStacks := c.filterTPStacks(stacks)
	if len(tpStacks) == 0 {
		return true, nil
	}

	if !c.forceMode {
		return false, fmt.Errorf("TerminationProtectionError: %s", c.joinStackNames(tpStacks))
	}

	return c.showTPConfirmation(tpStacks), nil
}

func (c *CdkStackConfirmer) confirmDeletion(stacks []cdk.StackInfo) bool {
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

func (c *CdkStackConfirmer) filterTPStacks(stacks []cdk.StackInfo) []cdk.StackInfo {
	var tp []cdk.StackInfo
	for _, s := range stacks {
		if s.TerminationProtection {
			tp = append(tp, s)
		}
	}
	return tp
}

func (c *CdkStackConfirmer) joinStackNames(stacks []cdk.StackInfo) string {
	names := make([]string, len(stacks))
	for i, s := range stacks {
		names[i] = s.StackName
	}
	return strings.Join(names, ", ")
}
