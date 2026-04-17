package runner

import (
	"context"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
	LookPath(name string) (string, error)
}

type OSRunner struct{}

func (r *OSRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return string(out), err
		}
		return string(out), fmt.Errorf("command %q failed: %w", name, err)
	}
	return string(out), nil
}

func (r *OSRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
