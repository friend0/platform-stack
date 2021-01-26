package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestBuildAllIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no arguments", []string{"help", "build", "all"}, "stack-build-all-help.golden"},
		{"build all command", []string{"build", "all"}, "stack-build-all-no-args-unconfigured.golden"},
		{"build all command with component", []string{"-r=../../examples/basic", "build", "app"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fixture != "" {
				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
				result, err := cmd.CombinedOutput()
				if err != nil {
					t.Log(err)
				}
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join("stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}
		})
	}
}
