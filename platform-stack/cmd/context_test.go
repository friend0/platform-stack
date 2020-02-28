package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os/exec"
	"path"
	"reflect"
	"testing"
)

func TestContextIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"context help", []string{"help", "context"}, "stack-context-help.golden"},
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

func TestRunContextFunctionGenerator(t *testing.T) {
	// override exec, and then make sure to reset it. The output of `fakeKubectlCommand` is "minikube"
	execCommand = fakeKubectlCommand
	defer func() { execCommand = exec.Command }()

	var buf bytes.Buffer
	fn := runContextFunction("current-context", &buf)
	assert.Equal(t, reflect.TypeOf(fn).String(), "func() error")
	err := fn()
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, buf.String(), "minikube")
}

func TestRunContextCommandFunctionGenerator(t *testing.T) {
	// override exec, and then make sure to reset it. The output of `fakeKubectlCommand` is "minikube"
	execCommand = fakeKubectlCommand
	defer func() { execCommand = exec.Command }()

	var buf bytes.Buffer
	fn := runContextCommandFunction("current-context", &buf)
	assert.Equal(t, reflect.TypeOf(fn).String(), "func(*cobra.Command, []string) error")
	err := fn(&cobra.Command{}, []string{})
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, buf.String(), "minikube")
}
