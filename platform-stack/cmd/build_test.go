package cmd

import (
	"fmt"
	"os/exec"
	"path"
	"testing"
)

func TestCliArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"no arguments", []string{}, "no-args.golden"},
		{"build", []string{}, "build-no-args.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//dir, err := os.Getwd()
			//if err != nil {
			//	t.Fatal(err)
			//}

			fmt.Println("ATTEMPTED PATH TO STACK::", path.Join(".", "stack"))
			cmd := exec.Command(path.Join(".", "stack"), tt.args...)
			_, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}

			//if *update {
			//	writeFixture(t, tt.fixture, output)
			//}

			//actual := string(output)

			//expected := loadFixture(t, tt.fixture)

			//if !reflect.DeepEqual(actual, expected) {
			//	t.Fatalf("actual = %s, expected = %s", actual, expected)
			//}
		})
	}
}