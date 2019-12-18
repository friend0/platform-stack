package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildRedisCmd represents the buildRedis command
var buildRedisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Builds Redis components.",
	Long: `Builds Redis components.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Redis components")
		return buildRedisComponents()
	},
}


func buildRedisComponents() (err error) {
	fmt.Println("Redis refers to an image. Skipping.")
	fmt.Println("Build Redis Succeeded")
	return nil
}


func init() {
	buildCmd.AddCommand(buildRedisCmd)
}
