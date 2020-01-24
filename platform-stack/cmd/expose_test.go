package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestExposeCLI(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		//{"expose", []string{"-r=../../examples", "expose", "missingComponent", "80", "80"}, "","stack-expose-nonexistent.golden"},
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
				//if err != nil {
				//	t.Error(err)
				//}
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}
