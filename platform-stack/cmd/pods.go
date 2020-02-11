package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// podsCmd represents the pods command
var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "List running pods.",
	Long:  `List running pods.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initK8s()
	},
	RunE: pods,
}

func pods(cmd *cobra.Command, args []string) (err error) {

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

	api := clientset.CoreV1()

	podList, err := getPodsList(api, ns, label, field)
	printPods(podList)
	return nil
}

func getPodsList(api v12.CoreV1Interface, ns string, label, field []string) (list *v1.PodList, err error) {

	defaultLabel := viper.GetString("stack")

	labelSelect := ""
	if defaultLabel != "" {
		labelSelect = fmt.Sprintf("stack=%v", defaultLabel)
	}
	for _, elem := range label {
		labelSelect += elem
	}

	fieldSelect := ""
	for _, elem := range field {
		fieldSelect += elem
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelect,
		FieldSelector: fieldSelect,
	}

	pods, err := api.Pods(ns).List(listOptions)
	if err != nil {
		return pods, err
	}
	return pods, nil
}

// printPods prints metadata about the pods in the provided list. It also returns this result as a byte array.
func printPods(pods *v1.PodList) (result []byte) {
	template := "%-50s%-8v%-24v%-8v%v\n"
	fmt.Printf(template, "NAME", "READY", "NAMESPACE", "STATUS", "IMAGES")
	result = append(result, fmt.Sprintf(template, "NAME", "READY", "NAMESPACE", "STATUS", "IMAGES")...)
	for _, pod := range pods.Items {
		numContainersReady := 0
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersReady++
			}

			if container.State.Terminated != nil && container.State.Terminated.ExitCode == 0 {
				numContainersReady++
			}
		}
		images := make([]string, len(pod.Spec.Containers))
		for i, container := range pod.Spec.Containers {
			images[i] = container.Image
		}
		fmt.Printf(template,
			pod.Name,
			fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
			pod.Namespace,
			pod.Status.Phase,
			images)
		result = append(result, []byte(fmt.Sprintf(template,
			pod.Name,
			fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
			pod.Namespace,
			pod.Status.Phase,
			images))...)
	}
	return result
}

func init() {
	rootCmd.AddCommand(podsCmd)

	podsCmd.Flags().StringP("namespace", "n", "", "Namespace")
	podsCmd.Flags().StringSliceP("label", "l", []string{}, "Label selector")
	podsCmd.Flags().StringSliceP("field", "f", []string{}, "Field selector")

}
