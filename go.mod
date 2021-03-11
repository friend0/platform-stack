module github.com/altiscope/platform-stack

go 1.15

replace (
	github.com/altiscope/platform-stack/pkg => ./pkg
	github.com/altiscope/platform-stack/platform-stack/cmd => ./platform-stack/cmd
)

require (
	github.com/GoogleContainerTools/skaffold v1.20.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/gookit/color v1.2.4
	github.com/magiconair/properties v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.2
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
)
