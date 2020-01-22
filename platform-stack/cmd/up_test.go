package cmd

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestUpCLI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"up", []string{"-r=../../examples", "up"}, "stack-up-no-args.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.fixture != "" {
				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
				result, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}

func mockEnv(required string) string {
	return required
}

func TestParseComponentArgs(t *testing.T) {

	componentArgs := []string{"app", "db"}
	parsedComponentArgs, _ := parseComponentArgs(componentArgs)
	assert.Equal(t, []ComponentDescription{
		{Name:"app"}, {Name:"db"},
	}, parsedComponentArgs)
}

func TestGenerateEnvs(t *testing.T) {

	requiredEnvs := []string{"var1", "var2"}
	generatedEnvs, _ := generateEnvs(requiredEnvs, mockEnv)
	assert.Equal(t, generatedEnvs, []string{`var1="var1"`, `var2="var2"`})
}