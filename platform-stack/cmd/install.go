package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs dependencies needed to run stack commands.",
	Long: `Installs dependencies needed to run stack commands.`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("install called")
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
