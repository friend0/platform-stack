package main

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"github.com/altiscope/platform-go-server"
)

func main() {

	// Here we're using the AutomaticEnv capability, but you could call a custom viper init here
	viper.AutomaticEnv()

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

	if err := http.ListenAndServe(fmt.Sprintf("%v", viper.GetString("GO_SERVER_API_PORT")), s.Engine); err != nil {
		return fmt.Errorf("startup error: %v", err.Error())
	}

	return nil
}
