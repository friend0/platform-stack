package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "List running pods.",
	Long: `List running pods.`,
	RunE: pods,
}

func pods(cmd *cobra.Command, args []string) (err error){

	ns, _ := cmd.Flags().GetString("ns")
	label, _ := cmd.Flags().GetString("label")
	field, _ := cmd.Flags().GetString("field")

	api := clientset.CoreV1()

	podList, err := getPodsList(api, ns, label, field)
	printPods(podList)
	return nil
}


func getPodsList(api v12.CoreV1Interface, ns, label, field string) (list *v1.PodList, err error) {

	listOptions := metav1.ListOptions{
	LabelSelector: label,
	FieldSelector: field,
	}

	pods, err := api.Pods(ns).List(listOptions)
	if err != nil {
		return pods, err
	}
	return pods, nil
}


// printPods prints metadata about the pods in the provided list. It also returns this result as a byte array.
func printPods(pods *v1.PodList) (result []byte) {
	template := "%-50s%-8v%-8v\n"
	fmt.Printf(template, "NAME", "READY", "STATUS")
	result = append(result, fmt.Sprintf(template, "NAME", "READY", "STATUS")...)
	for _, pod := range pods.Items {
		numContainersReady := 0
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersReady++
			}
		}
		fmt.Printf(template,
			pod.Name,
			fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
			pod.Status.Phase)
		result = append(result, []byte(fmt.Sprintf(template,
			pod.Name,
			fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
			pod.Status.Phase))...)
	}
	return result
}

func init() {

	initK8s()
	rootCmd.AddCommand(podsCmd)

	podsCmd.Flags().StringP("namespace", "n", "default", "Namespace")
	podsCmd.Flags().StringP("label", "l", "", "Label selector")
	podsCmd.Flags().StringP("field", "f", "", "Field selector")

}
