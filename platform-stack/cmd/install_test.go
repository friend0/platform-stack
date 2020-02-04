package cmd

import (
	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
	"path"
	"testing"
	"os/exec"
)

func TestDependencyExists(t *testing.T) {
	failingTests := []string{
		"exit 1",
		"exit 127",
	}

	for _, tt := range failingTests {
		exists := dependencyExists(tt)
		assert.False(t, exists)
	}

	passingTests := []string{
		"exit 0",
	}

	for _, tt := range passingTests {
		exists := dependencyExists(tt)
		assert.True(t, exists)
	}

}

func TestInstallDependency(t *testing.T) {
	failingTests := [][]string{
		{},
		{
			"exit 0",
			"exit 1",
		},
		{
			"exit 1",
		},
	}

	for _, tt := range failingTests {
		err := installDependency(tt, "")
		assert.Error(t, err)
	}

	passingTests := [][]string{
		{
			"exit 0",
		},
		{
			"exit 0",
			"exit 0",
		},
	}

	for _, tt := range passingTests {
		err := installDependency(tt, "")
		assert.NoError(t, err)
	}

}