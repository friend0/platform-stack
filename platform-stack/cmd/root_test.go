package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"os"
	"path"
	"testing"
	"os/exec"
)

func TestRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"build command", []string{"help"}, "stack-help.golden"},
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

func TestKubectlConfigProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "minikube")
	os.Exit(0)
}

func fakeKubectlCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestKubectlConfigProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGenerateCommand(t *testing.T) {
	// override exec, and then make sure to reset it
	execCommand = fakeKubectlCommand
	defer func() { execCommand = exec.Command }()

	cmd, err := GenerateCommand(`kubectl config {{.Data}}`, map[string]interface{}{"Data": "current-context"})
	if err != nil {
		t.Fail()
		return
	}
	assert.True(t, cmd.Args[len(cmd.Args)-1] == "kubectl config current-context")
	out, _ := cmd.CombinedOutput()
	assert.True(t, string(out) == "minikube")
}