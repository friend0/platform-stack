package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

func TestHealthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name      string
		args      []string
		setupArgs string
		fixture   string
	}{
		{"expose", []string{"help", "health"}, "", "stack-health-help.golden"},
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

func TestPodHealth(t *testing.T) {

	api := fake.NewSimpleClientset(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-app-579f7cd745-t6fdg",
			Namespace: "testns",
			Labels: map[string]string{
				"tag": "testtag",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "172.1.0.3",
		},
	})

	podList, err := getPodsList(api.CoreV1(), "testns", []string{"tag=testtag"}, []string{})
	if err != nil {
		t.Error(err.Error())
	}
	healthOutput := podHealth(podList)
	golden.AssertBytes(t, []byte(strings.Join(healthOutput, "")), "stack-health-one-healthy.golden")

}

func TestPodHealthWithUnhealthy(t *testing.T) {

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
					Image:           "foo/bar",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"echo hello"},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	})

	podList, err := getPodsList(api.CoreV1(), "testns", []string{"stack=testapp"}, []string{})
	if err != nil {
		t.Error(err.Error())
	}
	healthOutput := podHealth(podList)
	golden.AssertBytes(t, []byte(strings.Join(healthOutput, "")), "stack-health-one-unhealthy.golden")

}

func TestWaitForStack(t *testing.T) {

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
					Image:           "foo/bar",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"echo hello"},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	})

	cobraCmd := cobra.Command{

	}

	cobraCmd.Flags().StringP("namespace", "n", "", "Namespace")
	cobraCmd.Flags().StringSliceP("label", "l", []string{}, "Label selectors")
	cobraCmd.Flags().StringSliceP("field", "f", []string{}, "Field selectors")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	go waitForStack(api.CoreV1(), &cobraCmd, ctx)

	select {
	case <-time.After(1 * time.Nanosecond):
		cancel()
	case <-ctx.Done():
		t.Fail()
	}

	assert.Equal(t, ctx.Err(), context.Canceled)
}

func TestWaitForStackTimeout(t *testing.T) {

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
					Image:           "foo/bar",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"echo hello"},
				},
			},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	})

	cobraCmd := cobra.Command{

	}

	cobraCmd.Flags().StringP("namespace", "n", "", "Namespace")
	cobraCmd.Flags().StringSliceP("label", "l", []string{}, "Label selectors")
	cobraCmd.Flags().StringSliceP("field", "f", []string{}, "Field selectors")

	_, err, ctx := waitForStackWithTimeout(api.CoreV1(), &cobraCmd, 0)

	assert.Error(t, err, "context deadline exceeded")
	assert.Equal(t, ctx.Err(), context.DeadlineExceeded)
}
