package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/altiscope/platform-go-server/pkg"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	s := pkg.NewServer()
	// add server dependencies here as they come online
	s.InitDependencies("client")
	defer s.Close()

	if err := http.ListenAndServe(fmt.Sprintf("%v", os.Getenv("GO_SERVER_API_PORT")), s.Engine); err != nil {
		return fmt.Errorf("startup error: %v", err.Error())
	}

	return nil
}
