package cmd

import (
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestEnvironmentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"environment help", []string{"help", "environment"}, "stack-environment-help.golden"},
		{"environment no config", []string{"environment"}, "stack-environment-no-config-error.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestGetCurrentEnvironment(t *testing.T) {
	env, err := getCurrentEnvironment([]EnvironmentDescription{
		{
			Name:    "testenv",
			Activation: ActivationDescription{
				Env:     "env=activationtest",
				Context: "testcontext",
			},
		},
	}, "testcontext", func(string) string {
		return "activationtest"
	})

	if err != nil {
		t.Fail()
		return
	}

	assert.True(t, env.Activation.Context == "testcontext")

}
