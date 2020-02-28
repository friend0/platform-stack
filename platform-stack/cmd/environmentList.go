package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"os"
)

// environmentListCmd represents the environmentList command
var environmentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured environments.",
	Long:  `List configured environments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(config.Environments) == 0 {
			return fmt.Errorf("no environments found - double check you are in a stack directory with configured environments")
		}
		currentEnvironmentDescription, err := getEnvironment()
		if err != nil {
			return err
		}
		color.SetOutput(os.Stdout)
		for _, environment := range config.Environments {
			res, _ := json.MarshalIndent(environment, "", "    ")
			if currentEnvironmentDescription.Name == environment.Name {
				color.Info.Println(string(res))
			} else {
				fmt.Println(string(res))
			}
		}
		return nil
	},
}

func init() {
	environmentCmd.AddCommand(environmentListCmd)
}
