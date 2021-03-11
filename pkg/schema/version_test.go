package schema

import (
	"fmt"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"testing"

	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/altiscope/platform-stack/pkg/schema/util"
)

const (
	minimalConfig = ``

	simpleConfigNoApiVersion = `
stack:
  name: app
environments:
  - name: local
    activation:
      context: minikube
components: []
`

	simpleConfig = `
apiVersion: stack/v0beta1
stack:
  name: app
environments:
  - name: local
    activation:
      context: minikube
components: []
`
	completeConfig = `
stack:
  name: app
environments:
  - name: local
    activation:
      context: minikube
components:
  - name: app
    exposable: true
    containers:
      - dockerfile: ./containers/app/Dockerfile
        context: ./containers/app
        image: stack-app
    manifests:
      - ./deployments/app.yaml
`
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		apiVersion  string
		description string
		config      string
		expected    util.VersionedConfig
		shouldErr   bool
	}{
		{
			apiVersion:  latest.Version,
			description: "Minimal config",
			config:      minimalConfig,
			expected: config(
				withNoComponents(),
			),
		},
		{
			apiVersion:  "",
			description: "ApiVersion not specified",
			config:      simpleConfigNoApiVersion,
			expected: config(
				withStackDescription("app"),
				withLocalEnvironment(),
				withStackDescription("app"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Simple config",
			config:      simpleConfig,
			expected: config(
				withStackDescription("app"),
				withLocalEnvironment(),
				withNoComponents(),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Complete config",
			config:      completeConfig,
			expected: config(
				withStackDescription("app"),
				withLocalEnvironment(),
				withBasicComponent(),
			),
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.SetupFakeKubernetesContext(api.Config{CurrentContext: "cluster1"})
			tmpDir := t.NewTempDir().
				Write(".stack-local.yaml", fmt.Sprintf("%s", test.config))

			fmt.Println(tmpDir)
			cfg, err := ParseConfig(tmpDir.Path(".stack-local.yaml"), true)
			// todo: add handling of defaults here
			t.CheckErrorAndDeepEqual(test.shouldErr, err, test.expected, cfg)
		})
	}
}

func config(ops ...func(*latest.StackConfig)) *latest.StackConfig {
	cfg := &latest.StackConfig{
		ApiVersion: latest.Version,
		Environments: []latest.EnvironmentDescription{},
		Components: []latest.ComponentDescription{},
		Stack: latest.StackDescription{},
	}
	for _, op := range ops {
		op(cfg)
	}
	return cfg
}

func withStackDescription(name string) func(stackConfig *latest.StackConfig) {
	return func(cfg *latest.StackConfig) {
		cfg.Stack.Name = name
	}
}

func withNoEnvironment() func(stackConfig *latest.StackConfig) {
	return func(cfg *latest.StackConfig) {
		cfg.Environments = []latest.EnvironmentDescription{}
	}
}

func withLocalEnvironment(ops ...func(stackConfig *latest.EnvironmentDescription)) func(stackConfig *latest.StackConfig) {
	return func(cfg *latest.StackConfig) {
		b := latest.EnvironmentDescription{Name: "local", Activation: latest.ActivationDescription{Context: "minikube"}}
		for _, op := range ops {
			op(&b)
		}
		cfg.Environments = []latest.EnvironmentDescription{b}
	}
}

func withNoComponents(ops ...func(stackConfig *latest.EnvironmentDescription)) func(stackConfig *latest.StackConfig) {
	return func(cfg *latest.StackConfig) {
		cfg.Components = []latest.ComponentDescription{}
	}
}

func withBasicComponent(ops ...func(stackConfig *latest.EnvironmentDescription)) func(stackConfig *latest.StackConfig) {
	return func(cfg *latest.StackConfig) {
		cfg.Components = []latest.ComponentDescription{
			latest.ComponentDescription{
				Name:              "app",
				RequiredVariables: []string{},
				Exposable:         true,
				Containers: []latest.ContainerDescription{
					{
						Dockerfile:   "./containers/app/Dockerfile",
						Context:      "./containers/app",
						Image:        "stack-app",
						Environments: []string{},

					},
				},
				Manifests: []string{"./deployments/app.yaml"},
				TemplateConfig: []string{},
			},
		}
	}
}

func TestUpgradeToNextVersion(t *testing.T) {
	for i, schemaVersion := range VersionList[0 : len(VersionList)-2] {
		from := schemaVersion
		to := VersionList[i+1]
		description := fmt.Sprintf("Upgrade from %s to %s", from.APIVersion, to.APIVersion)

		testutil.Run(t, description, func(t *testutil.T) {
			factory, _ := VersionList.Find(from.APIVersion)

			newer, err := factory().Upgrade()

			t.CheckNoError(err)
			t.CheckDeepEqual(to.APIVersion, newer.GetVersion())
		})
	}
}

func TestCantUpgradeFromLatestVersion(t *testing.T) {
	factory, present := VersionList.Find(latest.Version)
	testutil.CheckDeepEqual(t, true, present)

	_, err := factory().Upgrade()
	testutil.CheckError(t, true, err)
}

// todo: withEnvironment

// todo: withComponent
