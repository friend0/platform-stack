package v0beta1

import (
	"github.com/altiscope/platform-stack/pkg/schema/util"
)

const Version string = "stack/v0beta1"

func NewStackConfig() util.VersionedConfig {
	return new(StackConfig)
}

func (config *StackConfig) GetVersion() string {
	return Version
}


type StackConfig struct {
	ApiVersion   string                   `yaml:"apiVersion"`
	Components   []ComponentDescription   `yaml:"components"`
	Environments []EnvironmentDescription `yaml:"environments"`
	Stack        StackDescription         `yaml:"stack"`
}

type StackDescription struct {
	Name string
}

type ActivationDescription struct {
	ConfirmWithUser bool
	Env             string
	Context         string
}

type EnvironmentDescription struct {
	Name       string
	Activation ActivationDescription
}

type ComponentDescription struct {
	Name              string                 `json:"name"`
	RequiredVariables []string               `json:"required_variables" yaml:"requiredVariables"`
	Exposable         bool                   `json:"exposable"`
	Containers        []ContainerDescription `json:"containers"`
	Manifests         []string               `json:"manifests"`
	TemplateConfig    []string               `json:"template_config"`
}

type ContainerDescription struct {
	Dockerfile   string   `json:"dockerfile"`
	Context      string   `json:"context"`
	Image        string   `json:"image"`
	Environments []string `json:"environments"`
}

type ManifestDescription struct {
	Dockerfile string `json:"dockerfile"`
	Context    string `json:"context"`
	Image      string `json:"image"`
}

type Config struct {
	Components   []ComponentDescription
	Environments []EnvironmentDescription
	Stack        StackDescription
}