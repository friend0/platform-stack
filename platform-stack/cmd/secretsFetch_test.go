package cmd

import (
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
	//"fmt"
)

func TestSecretsFetchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"secretsFetchHelp", []string{"-r=../../examples/basic", "help", "secrets", "fetch"}, "", "stack-secrets-fetch-help.golden"},
		// TODO: Fix this
		//{"secretsFetch with secrets", []string{"-r=../../examples/basic", "secrets", "fetch", "-e", "local"}, "", "stack-secrets-fetch-local-secrets.golden"},
		//{"secretsFetch with no args", []string{"-r=../../examples/basic", "secrets", "fetch"}, "", "stack-secrets-fetch-no-args.golden"},
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
				//fmt.Println(cmd)
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
