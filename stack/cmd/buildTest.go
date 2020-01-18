package cmd

import (
	"os"
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
		{"one argument", []string{"ciao"}, "one-argument.golden"},
		{"multiple arguments", []string{"ciao", "hello"}, "multiple-arguments.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(path.Join(dir, "stack"), tt.args...)
			_, err = cmd.CombinedOutput()
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