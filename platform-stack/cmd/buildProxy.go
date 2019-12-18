package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)



// buildProxyCmd represents the buildProxy command
var buildProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Builds Proxy components.",
	Long: `Builds Proxy components.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("Building Proxy components.")
		return buildProxyComponents()
	},
}


func buildProxyComponents() (err error) {
	cmd, err := GenerateCommand(dockerBuildTemplate, DockerBuildRequest{
		Dockerfile: "./containers/cloud-socks-proxy/Dockerfile",
		Image: "cloud-socks-proxy",
		Tag: "latest",
		Context: "./containers/cloud-socks-proxy",
	})

	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	buildCmd.AddCommand(buildProxyCmd)
}
