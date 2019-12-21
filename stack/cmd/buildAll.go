package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// buildAllCmd represents the buildAll command
var buildAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Builds all components for the Logflights stack.",
	Long: `Builds all components for the Logflights stack.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building all components")
		if !confirmWithUser("You are about to build and tag ALL components") {
			return nil
		}
		return buildAllComponents()
	},
}


func containsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}


func buildAllComponents() (err error) {

	//componentBuildMap := map[string]func()error {
	//	"database": buildDatabaseComponents,
	//	//"datadog": buildDatadogComponents,
	//	"frontend": buildFrontendComponents,
	//	"logflights": buildLogflightsComponents,
	//	"proxy": buildProxyComponents,
	//	"redis": buildRedisComponents,
	//}
	//
	//for component, componentUpFunction := range componentBuildMap {
	//	fmt.Println("Building component ", component)
	//	err := componentUpFunction()
	//	if err != nil {
	//		fmt.Printf("Build %v component failed", component)
	//		return err
	//	}
	//}
	return nil
}

func init() {
	buildCmd.AddCommand(buildAllCmd)
}
