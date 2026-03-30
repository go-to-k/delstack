package cdk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

const (
	DefaultCdkOutDir = "cdk.out"
	cdkJsonFile      = "cdk.json"
)

type Synthesizer struct {
	CdkOutDir string
}

func NewSynthesizer() *Synthesizer {
	return &Synthesizer{
		CdkOutDir: DefaultCdkOutDir,
	}
}

func (s *Synthesizer) Synth(ctx context.Context, contextValues []string, appCommand string) error {
	if appCommand == "" {
		if _, err := os.Stat(cdkJsonFile); os.IsNotExist(err) {
			return fmt.Errorf("cdk.json not found in current directory")
		}
	}

	args := []string{"synth", "--quiet"}
	if appCommand != "" {
		args = append(args, "--app", appCommand)
	}
	for _, cv := range contextValues {
		args = append(args, "-c", cv)
	}

	cmd := exec.CommandContext(ctx, "npx", append([]string{"cdk"}, args...)...) //nolint:gosec // args are constructed from trusted CLI flags
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cdk synth failed: %w", err)
	}

	return nil
}
