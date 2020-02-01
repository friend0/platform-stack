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

	api := clientset.CoreV1()

	ns, _ := cmd.Flags().GetString("namespace")
	label, _ := cmd.Flags().GetStringSlice("label")
	field, _ := cmd.Flags().GetStringSlice("field")

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

	podList, err := getPodsList(api, ns, labelSelect, fieldSelect)
	fmt.Println(podHealth(podList))
	return nil
}

// podHealth generates a report string given an input PodList
func podHealth(pods *v1.PodList) (output string) {

	for _, pod := range pods.Items {
		healthy := true
		numContainersHealthy := 0

		podDetailHeader := fmt.Sprintf("\n\tPod Details `%v`\n", pod.Name)
		podDetailOutput := podDetailHeader
		podDetailOutput += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(podDetailHeader)))

		for _, condition := range pod.Status.Conditions {
			podDetailOutput += fmt.Sprintf("\t%v: %v\n", condition.Type, condition.Status)
		}

		// check container numbers
		for _, container := range pod.Status.ContainerStatuses {
			if container.Ready {
				numContainersHealthy++
			}

			if container.State.Waiting != nil || container.State.Terminated != nil {
				containerDetailHeader := fmt.Sprintf("\n\tContainer Details `%v`\n", container.Name)
				podDetailOutput += containerDetailHeader
				podDetailOutput += fmt.Sprintf("\t%v\n", strings.Repeat("=", utf8.RuneCountInString(containerDetailHeader)))
			}
			if container.State.Waiting != nil {
				podDetailOutput += fmt.Sprintf("\tContainer Waiting: %v\n", container.State.Waiting.Message)
				healthy = false
			}
			if container.State.Terminated != nil {
				if container.State.Terminated.ExitCode == 0 {
					numContainersHealthy++
				} else {
					podDetailOutput += fmt.Sprintf("\tContainer Terminated with non-zero ExitCode: %v: %v\n", container.State.Terminated.ExitCode, container.State.Terminated.Message)
					healthy = false
				}
			}
		}

		if numContainersHealthy != len(pod.Spec.Containers) {
			healthy = false
		}

		if healthy {
			output += fmt.Sprintf("✔️  %v in namespace `%v` is healthy\n", pod.Name, pod.Namespace)
		} else {
			output += fmt.Sprintf("✖️  %v in namespace `%v` is not healthy\n", pod.Name, pod.Namespace) + podDetailOutput
		}
	}
	return output
}

func init() {
	rootCmd.AddCommand(healthCmd)

	initK8s()

	healthCmd.Flags().StringP("namespace", "n", "", "Namespace")
	healthCmd.Flags().StringSliceP("label", "l", []string{}, "Label selectors")
	healthCmd.Flags().StringSliceP("field", "f", []string{}, "Field selectors")
}
