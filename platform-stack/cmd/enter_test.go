package cmd

import (
	"fmt"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestEnterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"enter", []string{"-r=../../examples", "help", "enter"}, "", "stack-enter-noargs.golden"},
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

func TestEnterContainerCommand(t *testing.T) {

	api := fake.NewSimpleClientset(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-app-579f7cd745-t6fdg",
			Namespace: "testns",
			Labels: map[string]string{
				"stack": "testapp",
			},
		},

		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "172.1.0.3",
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "Back-off 5m0s restarting failed container=foo/bar pod=tls-app-579f7cd745-t6fdg_default(81cf37f3-3dff-11ea-b7c5-025000000001)",
						},
					},
				},
			},
			Conditions: []v1.PodCondition{
				{
					Type:   "Initialized",
					Status: v1.ConditionTrue,
				},
				{
					Type:   "Ready",
					Status: v1.ConditionFalse,
				},
				{
					Type:   "ContainerReady",
					Status: v1.ConditionFalse,
				},
				{
					Type:   "PodScheduled",
					Status: v1.ConditionTrue,
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "tls-app",
					Image:           "alpine:latest",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"cat"},
					Stdin:           true,
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	})

	podList, err := getPodsList(api.CoreV1(), "testns", []string{"stack=testapp"}, []string{})
	if err != nil {
		t.Fail()
		return
	}

	fmt.Println(podList)
	var targetPod *v1.Pod
	if len(podList.Items) < 1 {
		t.Fail()
	} else {
		targetPod = &podList.Items[0]
	}

	enterCommand, err := enterContainerCommand(targetPod, "tls-app", "bin/sh")
	golden.Assert(t, strings.Join(enterCommand.Args, " "), "stack-enter-command.golden")
}
