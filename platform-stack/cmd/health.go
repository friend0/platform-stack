package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

// podsCmd represents the pods command
var healthCmd = &cobra.Command{
	Use:   "pods",
	Short: "Get the health of the stack.",
	Long:  `List running pods.`,
	RunE:  health,
}

func health(cmd *cobra.Command, args []string) (err error) {

	ns, _ := cmd.Flags().GetString("ns")
	label, _ := cmd.Flags().GetString("label")
	field, _ := cmd.Flags().GetString("field")

	api := clientset.CoreV1()

	podList, err := getPodsList(api, ns, label, field)
	podHealth(podList)
	return nil
}

func podHealth(pods *v1.PodList) {
	template := "%-50s%-8v%-8v\n"
	fmt.Printf(template, "NAME", "READY", "STATUS")
	//result = append(result, fmt.Sprintf(template, "NAME", "READY", "STATUS")...)
	for _, pod := range pods.Items {
		numContainersReady := 0
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersReady++
			}
		}
		fmt.Println("POD CONDITIONS: ", pod.Status.Conditions)
		fmt.Printf(template,
			pod.Name,
			fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
			pod.Status.Phase)
		//result = append(result, []byte(fmt.Sprintf(template,
		//	pod.Name,
		//	fmt.Sprintf("%v/%v", numContainersReady, len(pod.Spec.Containers)),
		//	pod.Status.Phase))...)
	}
}

func init() {
	rootCmd.AddCommand(healthCmd)

	initK8s()

	healthCmd.Flags().StringP("namespace", "n", "default", "Namespace")
	healthCmd.Flags().StringP("label", "l", "", "Label selector")
	healthCmd.Flags().StringP("field", "f", "", "Field selector")

}
