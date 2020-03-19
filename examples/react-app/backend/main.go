package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {

	s := NewServer()
	defer s.Close()

	addr := GetEnv("BACKEND_API_PORT", "5001")
	log.Printf("Server listening on port %v...\n", addr)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", addr), s.Router); err != nil {
		log.Fatal("There was an error starting the server", err)
	}

	return nil
}
