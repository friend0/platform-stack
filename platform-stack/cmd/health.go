package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"strings"
	"unicode/utf8"
)

// podsCmd represents the pods command
var healthCmd = &cobra.Command{
	Use:   "health",
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

	for _, pod := range pods.Items {
		healthy := true
		detail := ""
		numContainersReady := 0

		podDetailHeader := fmt.Sprintf("\n\tPod Details `%v`\n", pod.Name)
		detail += podDetailHeader
		detail += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(podDetailHeader)))

		for _, condition := range pod.Status.Conditions {
			detail += fmt.Sprintf("\t%v: %v\n", condition.Type, condition.Status)
		}

		// check container numbers
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersReady++
			}

			if container.State.Waiting != nil || container.State.Terminated != nil {
				containerDetailHeader := fmt.Sprintf("\n\tContainer Details `%v`\n", container.Name)
				detail += containerDetailHeader
				detail += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(containerDetailHeader)))
			}
			if container.State.Waiting != nil {
				detail += fmt.Sprintf("\tContainer %v Waiting: %v\n", container.Name, container.State.Waiting.Message)
				healthy = false
			}
			if container.State.Terminated != nil{
				detail += fmt.Sprintf("\tContainer %v Terminated: %v\n", container.Name, container.State.Terminated.Message)
				healthy = false
			}
		}

		if numContainersReady != len(pod.Spec.Containers) {
			healthy = false
		}


		if healthy {
			fmt.Printf("✔️  %v is healthy\n", pod.Name)
		} else {
			fmt.Printf("✖️  %v is not healthy", pod.Name)
			// check pod conditions'

			fmt.Println(detail)
		}
	}
}

func init() {
	rootCmd.AddCommand(healthCmd)

	initK8s()

	healthCmd.Flags().StringP("namespace", "n", "default", "Namespace")
	healthCmd.Flags().StringP("label", "l", "", "Label selector")
	healthCmd.Flags().StringP("field", "f", "", "Field selector")

}
