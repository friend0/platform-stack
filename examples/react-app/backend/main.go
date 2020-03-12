package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	s := NewServer()
	s.InitDependencies("database", "logger")
	defer s.Close()

	addr := GetEnv("API_PORT", ":5001")
	if err := http.ListenAndServe(fmt.Sprintf("%v", addr), s.Router); err != nil {
		s.Log.Fatal("There was an error starting the server", err)
	}

	return nil
}
