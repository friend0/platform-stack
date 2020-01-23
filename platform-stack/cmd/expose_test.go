package cmd

import (
	"fmt"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestExposeCLI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setupArgs string
		fixture string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if len(tt.setupArgs) > 1 {
				cmd := exec.Command("sh", "-c", tt.setupArgs)
				result, err := cmd.CombinedOutput()
				if err != nil {
					t.Error(err)
				}
				fmt.Println(string(result))
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