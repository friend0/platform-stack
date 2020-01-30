package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetString("field")

	api := clientset.CoreV1()

	defaultLabel := viper.GetString("stack")

	labelSelect := ""
	if defaultLabel != "" {
		labelSelect = fmt.Sprintf("stack=%v", defaultLabel)
	}
	for _, elem := range label {
		labelSelect += elem
	}
	
	podList, err := getPodsList(api, ns, labelSelect, field)
	fmt.Println(podHealth(podList))
	return nil
}

// podHealth generates a report string given an input PodList
func podHealth(pods *v1.PodList) (output string) {


	for _, pod := range pods.Items {
		healthy := true
		numContainersReady := 0

		podDetailHeader := fmt.Sprintf("\n\tPod Details `%v`\n", pod.Name)
		output += podDetailHeader
		output += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(podDetailHeader)))

		for _, condition := range pod.Status.Conditions {
			output += fmt.Sprintf("\t%v: %v\n", condition.Type, condition.Status)
		}

		// check container numbers
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersReady++
			}

			if container.State.Waiting != nil || container.State.Terminated != nil {
				containerDetailHeader := fmt.Sprintf("\n\tContainer Details `%v`\n", container.Name)
				output += containerDetailHeader
				output += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(containerDetailHeader)))
			}
			if container.State.Waiting != nil {
				output += fmt.Sprintf("\tContainer %v Waiting: %v\n", container.Name, container.State.Waiting.Message)
				healthy = false
			}
			if container.State.Terminated != nil{
				output += fmt.Sprintf("\tContainer %v Terminated: %v\n", container.Name, container.State.Terminated.Message)
				healthy = false
			}
		}

		if numContainersReady != len(pod.Spec.Containers) {
			healthy = false
		}


		if healthy {
			output = fmt.Sprintf("✔️  %v is healthy\n", pod.Name)
		} else {
			output = fmt.Sprintf("✖️  %v is not healthy\n", pod.Name) + output
		}
	}
	return output
}

func init() {
	rootCmd.AddCommand(healthCmd)

	initK8s()

	healthCmd.Flags().StringP("namespace", "n", "default", "Namespace")
	healthCmd.Flags().StringSliceP("label", "l", []string{}, "Label selector")
	healthCmd.Flags().StringP("field", "f", "", "Field selector")
}
