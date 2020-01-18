module github.com/altiscope/platform-stack

go 1.12

replace github.com/altiscope/platform-stack/platform-stack/cmd => ./platform-stack/cmd

require (
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
)
