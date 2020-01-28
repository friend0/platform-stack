package cmd

import (
	"gotest.tools/v3/golden"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)


//func TestHealthCLI(t *testing.T) {
//	tests := []struct {
//		name      string
//		args      []string
//		setupArgs string
//		fixture   string
//	}{
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			if tt.setupArgs != "" {
//				cmd := exec.Command("sh", "-c", tt.setupArgs)
//				_, err := cmd.CombinedOutput()
//				if err != nil {
//					t.Error(err)
//				}
//			}
//
//			if tt.fixture != "" {
//				cmd := exec.Command(path.Join(".", "stack"), tt.args...)
//				result, _ := cmd.CombinedOutput()
//				//if err != nil {
//				//	t.Error(err)
//				//}
//				golden.AssertBytes(t, result, tt.fixture)
//			} else {
//				result := icmd.RunCmd(icmd.Command(path.Join(".", "stack"), tt.args...))
//				result.Assert(t, icmd.Success)
//			}
//
//		})
//	}
//}

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

	podList, err := getPodsList(api.CoreV1(), "testns", "tag=testtag", "")
	if err != nil {
		t.Error(err.Error())
	}
	healthOutput := podHealth(podList)
	golden.AssertBytes(t, []byte(healthOutput), "stack-pod-all-healthy.golden")

}
