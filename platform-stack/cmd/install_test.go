package cmd

import (
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"path"
	"testing"
	"os/exec"
)

func TestDependencyExists(t *testing.T) {
	failingTests := []string{
		"exit 1",
		"exit 127",
	}

	for _, tt := range failingTests {
		exists := dependencyExists(tt)
		assert.False(t, exists)
	}

	passingTests := []string{
		"exit 0",
	}

	for _, tt := range passingTests {
		exists := dependencyExists(tt)
		assert.True(t, exists)
	}

}

func TestInstallDependency(t *testing.T) {
	failingTests := [][]string{
		{},
		{
			"exit 0",
			"exit 1",
		},
		{
			"exit 1",
		},
	}

	for _, tt := range failingTests {
		err := installDependency(tt, "")
		assert.Error(t, err)
	}

	passingTests := [][]string{
		{
			"exit 0",
		},
		{
			"exit 0",
			"exit 0",
		},
	}

	for _, tt := range passingTests {
		err := installDependency(tt, "")
		assert.NoError(t, err)
	}

}

func TestInstallCLI(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"install help", []string{"help", "install"}, "", "stack-install-help.golden"},
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
				golden.AssertBytes(t, result, tt.fixture)
			} else {
				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
				result.Assert(t, icmd.Success)
			}

		})
	}
}
