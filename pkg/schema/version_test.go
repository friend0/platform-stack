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

	simpleConfig = `
apiVersion: stack/v0beta1
stack:
  name: app
environments:
  - name: local
    activation:
      context: cluster1
components: []
`
	completeConfig = `
apiVersion: stack/v0beta1
stack:
  name: app
environments:
  - name: local
    activation:
      context: cluster1
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
				withLocalEnvironment(
				),
				//withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  "stack/v1alpha1",
			description: "Old minimal config",
			config:      minimalConfig,
			expected: config(
				withLocalEnvironment(
					//withGitTagger(),
				),
				//withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Simple config",
			config:      simpleConfig,
			expected: config(
				withLocalEnvironment(
					//withGitTagger(),
					//withDockerArtifact("example", ".", "Dockerfile"),
				),
				//withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Complete config",
			config:      completeConfig,
			expected: config(
				withLocalEnvironment(
					//withShaTagger(),
					//withDockerArtifact("image1", "./examples/app1", "Dockerfile.dev"),
					//withBazelArtifact("image2", "./examples/app2", "//:example.tar"),
				),
				//withKubectlDeploy("dep.yaml", "svc.yaml"),
			),
		},
		{
			apiVersion:  "",
			description: "ApiVersion not specified",
			config:      minimalConfig,
			shouldErr:   true,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.SetupFakeKubernetesContext(api.Config{CurrentContext: "cluster1"})

			tmpDir := t.NewTempDir().
				Write(".stack.yaml", fmt.Sprintf("apiVersion: %s\nkind: Config\n%s", test.apiVersion, test.config))

			cfg, err := ParseConfig(tmpDir.Path(".stack.yaml"), true)
			// todo: add handling of defaults here
			t.CheckErrorAndDeepEqual(test.shouldErr, err, test.expected, cfg)
		})
	}
}

func config(ops ...func(*latest.StackConfig)) *latest.StackConfig {
	cfg := &latest.StackConfig{ApiVersion: latest.Version, Stack: latest.StackDescription{Name: "testApp"}}
	for _, op := range ops {
		op(cfg)
	}
	return cfg
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