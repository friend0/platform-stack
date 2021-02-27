package cmd

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetPodList(t *testing.T) {

	api := fake.NewSimpleClientset(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "stack/v1alpha1",
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
	podListString := podList.String()
	golden.AssertBytes(t, []byte(podListString), "stack-get-pods-list.golden")

}

func TestPrintPods(t *testing.T) {

	api := fake.NewSimpleClientset(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-app-579f7cd745-t6fdg",
			Namespace: "default",
			Labels: map[string]string{
				"tag": "",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "172.1.0.3",
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

	pods, _ := api.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	assert.True(t, len(pods.Items) == 1)

	var buf bytes.Buffer
	_, err := printPodList(pods, &buf)
	if err != nil {
		t.Fail()
	}
	golden.AssertBytes(t, buf.Bytes(), "stack-print-pods-list.golden")
}
