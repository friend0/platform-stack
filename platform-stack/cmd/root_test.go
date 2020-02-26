package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"path"
	"testing"
	"os/exec"
)

func TestRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"build command", []string{"help"}, "stack-help.golden"},
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