package cmd

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
)

func TestDownIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"down", []string{"-r=../../examples/basic", "down"}, "", "stack-down-no-args-none-running.golden"},
		{"down with running", []string{"-r=../../examples/basic", "down", "app", "config"}, "stack -r=../../examples/basic up", "stack-down-multiple-args-with-running.golden"},
		{"down with running", []string{"-r=../../examples/basic", "down"}, "stack -r=../../examples/basic down; stack -r=../../examples/basic up", "stack-down-no-args-with-running.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if len(tt.setupArgs) > 1 {
				cmd := exec.Command("sh", "-c", tt.setupArgs)
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, "ENV=local")
				_, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
			}

			if tt.fixture != "" {
				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, "ENV=local")
				result, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
				golden.Assert(t, string(result), tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}
