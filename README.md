# Stack
Stack is the Platform deployment CLI.

Add a minimal `.stack.yaml` configuration to your project, and Stack will help make the development and deployment 
of your service easier. 

## Install

You can get started with stack right away by downloading the latest release from [github](https://github.com/altiscope/platform-stack/releases).

If you have go installed, you can also build stack from source and put in onto your path with `go build -o /usr/local/bin/stack -v ../platform-stack/main.go`.

Once stack is available, system dependencies can be installed by running `stack install`.  

To develop against a local kubernetes cluster, docker-desktop is the simplest path forward. 
For this, you'll need to follow the install steps described [here](https://docs.docker.com/docker-for-mac/install/).


## Core Commands

Stack provides a set of core commands - `build`, `up`, and `down`:

- `build` provides simple building and tagging.

- `up` will take templated k8s definitions, hydrate them with config, and apply them to the configured cluster (todo: context aware).

- `down` will terminate those applied objects. 

Stack assumes that your project maintains a containers directory for container definitions, and a `deployments` 
directory for kubernetes object definitions.

These directories can be configured in the project's `.stack.yaml` as build_directory, and deployment_directory 
respecitvely.


Stack uses a configurable set of components along with the above directory configurations to properly scope commands.
 

Components should be configured in `.stack.yaml` as a list of component description objects. The name of the component
should correspond to it's name in the deployments directory. For each component, you may specify 
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


The `up` command uses the name of configured components to locate kubernetes manifests. In the list of descriptions 
above, `app` should have a corresponding `app.yaml` in the deployments directory.

## Development

Stack also provides tools useful for development, and debugging.

- `expose` will port forward between a running pod and the local machine