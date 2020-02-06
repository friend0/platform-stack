package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"reflect"
	"testing"
)

func TestParseDependencyVersionOverrides(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		input          []string
		expectedOutput map[string]string
	}{
		{[]string{""}, map[string]string{}},
		{[]string{"k=v1.17.1"}, map[string]string{"k": "v1.17.1"}},
		{[]string{"a=v1.0.1", "b=1.0"}, map[string]string{"a": "v1.0.1", "b": "1.0"}},
	}

	for _, tt := range tests {
		output, err := parseDependencyVersionOverrides(tt.input)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(output)
		assert.True(t, reflect.DeepEqual(tt.expectedOutput, output))
	}

	errTests := []struct {
		input []string
	}{
		{[]string{"a="}},
		{[]string{"a=", "b=1.0"}},
	}

	for _, tt := range errTests {
		_, err := parseDependencyVersionOverrides(tt.input)
		assert.Error(t, err)
	}

}

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

func TestInstallIntegration(t *testing.T) {
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
