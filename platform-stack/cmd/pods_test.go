package cmd

import (
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
	})

	pods, _ := api.CoreV1().Pods("default").List(metav1.ListOptions{})
	assert.True(t, len(pods.Items) == 1)

	printBytes := printPods(pods)
	golden.AssertBytes(t, printBytes, "stack-print-pods-list.golden")
}
