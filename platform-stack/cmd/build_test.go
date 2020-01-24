package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestBuildCLI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no arguments", []string{}, "stack-no-args.golden"},
		{"build command", []string{"build"}, "stack-build-no-args.golden"},
		{"build command with component", []string{"-r=../../examples", "build", "app"}, ""},
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
