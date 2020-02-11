package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestLogsCLI(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"logs help", []string{"help", "logs"}, "", "stack-logs-help.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setupArgs != "" {
				cmd := exec.Command("sh", "-c", tt.setupArgs)
				_, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
			}

			if tt.fixture != "" {
				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
				result, _ := cmd.CombinedOutput()
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}
