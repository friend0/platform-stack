package cmd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// todo: build latest binary
	os.Exit(m.Run())
}