package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
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
		{"down", []string{"-r=../../examples", "down"}, "", "stack-down-no-args-none-running.golden"},
		{"down with running", []string{"-r=../../examples", "down"}, "stack -r=../../examples up", "stack-down-no-args-with-running.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if len(tt.setupArgs) > 1 {
				cmd := exec.Command("sh", "-c", tt.setupArgs)
				_, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
			}

			if tt.fixture != "" {
				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
				result, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}
