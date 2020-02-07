package cmd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	// setup
	os.Exit(m.Run())
	// teardown
}
