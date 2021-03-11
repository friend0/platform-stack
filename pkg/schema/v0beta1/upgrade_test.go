package v0beta1

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/yaml"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"testing"
)

func TestUpgrade_local(t *testing.T) {
	yaml := `stack:
  name: app
environments:
  - name: local
    activation:
      env: ENV=local
      context: docker-desktop || minikube || microk8s
  - name: ci
    activation:
      env: ENV=ci
      context: minikube
  - name: staging
    activation:
      env: ENV=staging
      context: platform-stg-blue
components:
  - name: app
    exposable: true
    containers:
      - dockerfile: ./containers/app/Dockerfile
        context: ./containers/app
        image: stack-app
        environments:
          - local
    manifests:
      - ./deployments/app.yaml
    requiredVariables:
      - RSA_KEY
    templateConfig: []
`
	expected := `apiVersion: stack/v1alpha1
stack:
  name: app
environments:
  - name: local
    activation:
      env: ENV=local
      context: docker-desktop || minikube || microk8s
  - name: ci
    activation:
      env: ENV=ci
      context: minikube
  - name: staging
    activation:
      env: ENV=staging
      context: platform-stg-blue
components:
  - name: app
    exposable: true
    containers:
      - dockerfile: ./containers/app/Dockerfile
        context: ./containers/app
        image: stack-app
        environments:
          - local
    manifests:
      - ./deployments/app.yaml
    environments: []
    requiredVariables:
      - RSA_KEY
    templateConfig: []
`
	verifyUpgrade(t, yaml, expected)
}

func verifyUpgrade(t *testing.T, input, output string) {
	config := NewStackConfig()
	err := yaml.UnmarshalStrict([]byte(input), config)
	testutil.CheckErrorAndDeepEqual(t, false, err, Version, config.GetVersion())

	upgraded, err := config.Upgrade()
	testutil.CheckError(t, false, err)

	expected := latest.NewStackConfig()
	err = yaml.UnmarshalStrict([]byte(output), expected)

	testutil.CheckErrorAndDeepEqual(t, false, err, expected, upgraded)
}
