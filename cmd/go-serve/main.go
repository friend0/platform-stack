package main

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"github.com/altiscope/platform-go-server"
)

func main() {

	viper.SetDefault("API_DB_HOST", "postgres")
	viper.SetDefault("API_DB_PORT", "5432")
	viper.SetDefault("API_DB_USERNAME", "postgres")
	viper.SetDefault("API_DB_PASSWORD", "postgres")
	viper.SetDefault("API_DB_NAME", "postgres")
	viper.SetDefault("API_DB_MODE", "disable")

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	s := server.NewServer()
	// add server dependencies here as they come online
	s.InitDependencies("client")
	defer s.Close()

	if err := http.ListenAndServe(fmt.Sprintf("%v", os.Getenv("GO_SERVER_API_PORT")), s.Engine); err != nil {
		return fmt.Errorf("startup error: %v", err.Error())
	}

	return nil
}
