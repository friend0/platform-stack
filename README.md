# Stack
Stack is the Platform deployment CLI.

Stack is a generalized CLI for seamless test, development, and debugging across environments.
Currently for local development only, stack has the potential to minimize dev/prod deltas, and to give developers
a powerful set of tools that make them more productive. 

Stack makes deployments simple by making a few organizational assumptions about your service:

1) Containers are defined in a single directory
2) Deployment manifests are defined in another directory - these are sometimes referred to within these docs as
components
3) A configuration file is present defining the above directories, as well as the components that will be managed
as a part of the stack

Over time, stack can become less opinionated about how your project is organized - as of now, no free lunch!

## Getting Started

### Install

Option 1: Install `jq` with `brew install jq`, then run the install script `install.sh`
Option 2: Navigate to https://github.com/altiscope/platform-stack/releases and take the latest.
Next, put the appropriate binary onto your path renamed as `stack`
Option 3: Build from source `go build -o /usr/local/bin/stack -v ../platform-stack/main.go`

Once stack is available, system dependencies can be installed by running `stack install`.  

To develop against a local kubernetes cluster, docker-desktop is the simplest path forward. 
For this, you'll need to follow the install steps described [here](https://docs.docker.com/docker-for-mac/install/).

### Project Setup and Configuration

Stack assumes that your project maintains a containers directory for container definitions, and a `deployments` 
directory for kubernetes object definitions.

These directories can be configured in the project's configuration file as `build_directory`, and `deployment_directory` 
respectively.


Stack uses a configurable set of components along with the above directory configurations to properly scope commands.
 

Components should be defined in the configuration file as a list of component description objects. The name
must correspond to it's name in the deployments directory. For each component, you may specify 
required environment variables, or whether or not a component is able to be exposed.

    components:
      - name: config
        requiredVariables:
          - ENV
      - name: app
        requiredVariables:
          - PWD
          - HOME
        exposable: true
        
### Running

The stack CLI assumes the present working directory is the root project directory, and that a configuration file 
exists. Alternately, you can provide the desired root directory (with configuration file) by setting the `project_directory` flag on the root stack command.
You can run using a specific configuration file in that directory setting the `configuration_file` flag on the root stack command.

In a directory with project components defined and configured, you may run `stack build` then `stack up` to bring up the
latest version of the stack.

## Core Commands

Stack provides a set of core commands - `build`, `up`, and `down`:

- `build` provides simple building and tagging.

- `up` will take templated k8s definitions, hydrate them with config, and apply them to the configured cluster (todo: context aware).

- `down` will terminate those applied objects. 

The `up` command uses the name of configured components to locate kubernetes manifests. In the list of descriptions 
above, `app` should have a corresponding `app.yaml` in the deployments directory.



## Development

Stack also provides tools useful for development, and debugging.

- `expose` will port forward between a running pod and the local machine
- `logs` will display the logging output of a running pod
