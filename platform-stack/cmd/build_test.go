package cmd

import (
	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"testing"
)

func TestBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"build command", []string{"help", "build"}, "stack-build-help.golden"},
		{"build command", []string{"build"}, "stack-build-no-args.golden"},
		{"build command with component", []string{"-r=../../examples/basic", "build", "app"}, ""},
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

func TestBuildForCurrentEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		stackEnv string
		cd       latest.ContainerDescription
		res      bool
	}{
		{"build no environments given", "local", latest.ContainerDescription{}, true},
		{"build single environment given, no match", "local", latest.ContainerDescription{
			Environments: []string{"remote"},
		}, false},
		{"build single environment given, match", "local", latest.ContainerDescription{
			Environments: []string{"local"},
		}, true},
		{"build multiple environments given, no match", "ci", latest.ContainerDescription{
			Environments: []string{"local", "remote"},
		}, false},
		{"build multiple environments given, match", "remote", latest.ContainerDescription{
			Environments: []string{"local", "remote"},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := buildForCurrentEnvironment(tt.cd, tt.stackEnv)
			assert.Equal(t, res, tt.res)

		})
	}
}
