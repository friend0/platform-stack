module github.com/altiscope/platform-stack

go 1.12

replace (
	github.com/altiscope/platform-stack/pkg => ./pkg
	github.com/altiscope/platform-stack/platform-stack/cmd => ./platform-stack/cmd
	github.com/altiscope/platform-stack/schema => ./schema
)

require (
	github.com/GoogleContainerTools/skaffold v1.7.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/gookit/color v1.2.2
	github.com/magiconair/properties v1.8.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.7
	gotest.tools/v3 v3.0.2
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20200117235808-5f6fbceb4c31 // indirect
)
