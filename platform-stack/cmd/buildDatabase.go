package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildDatabaseCmd represents the buildDatabase command
var buildDatabaseCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Builds LAANC database objects.",
	Long:    `Builds LAANC database objects.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Build Database Components")
		return buildDatabaseComponents()
	},
}

func buildDatabaseComponents() (err error) {
	fmt.Println("Database refers to an image. Skipping.")
	fmt.Println("Build Database Succeeded")
	return nil
}

func init() {
	buildCmd.AddCommand(buildDatabaseCmd)
}