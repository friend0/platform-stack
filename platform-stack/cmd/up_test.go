package cmd

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/magiconair/properties/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
)

func TestUpCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"up", []string{"-r=../../examples/react-app", "up", "config", "backend"}, "stack-up-react-app.golden"},
		{"up", []string{"-r=../../examples/basic", "up", "app"}, "stack-up-app.golden"},
		{"up", []string{"-r=../../examples/basic", "up"}, "stack-up-no-args.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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

func TestParseComponentArgs(t *testing.T) {
	componentArgs := []string{"app", "db"}
	configuredComponents := []latest.ComponentDescription{
		{Name: "app"},
		{Name: "db"},
	}
	parsedComponentArgs, _ := parseComponentArgs(componentArgs, configuredComponents)
	assert.Equal(t, configuredComponents, parsedComponentArgs)
}

func mockEnv(required string) string {
	return required
}

func TestGenerateEnvs(t *testing.T) {

	requiredEnvs := []string{"var1", "var2"}
	generatedEnvs, _ := generateEnvs(requiredEnvs, mockEnv)
	assert.Equal(t, generatedEnvs, []string{`var1="var1"`, `var2="var2"`})
}
