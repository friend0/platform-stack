module github.com/altiscope/platform-stack

go 1.12

replace github.com/altiscope/platform-stack/platform-stack/cmd => ./platform-stack/cmd

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gookit/color v1.2.2
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	gotest.tools/v3 v3.0.0
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20200117235808-5f6fbceb4c31 // indirect
)
