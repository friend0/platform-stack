
# ☰ Stack

Stack is a tool for defining and running multi-object Kubernetes applications. 
With Stack, you use a configuration file to define the services that make up your application. 
Then, with a few simple commands, you create and start all the services from your configuration. 

Stack is a generalized CLI for seamless test, development, and debugging across environments.
Currently for local development only, stack has the potential to minimize dev/prod deltas, and to give developers
a powerful set of tools for developing and maintaining services.


## 🚀 Getting Started

### Installation

Option 1: Install `jq` with `brew install jq`, then run the install script `install.sh`
You will need to export a github personal access token as GIT_TOKEN `export GIT_TOKEN=<GENERATED_TOKEN_HERE>` [see here](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line).

Option 2: Navigate to https://github.com/altiscope/platform-stack/releases and take the latest.
Next, put the appropriate binary onto your path renamed as `stack`
Option 3: Build from source `go build -o /usr/local/bin/stack -v ../platform-stack/main.go`

Once stack is available, system dependencies can be installed by running `stack install`.  

To develop against a local kubernetes cluster, docker-desktop is the simplest path forward. 
For this, you'll need to follow the install steps described [here](https://docs.docker.com/docker-for-mac/install/).

### The Stack Configuration File

The Stack CLI requires a project configuration file to properly interface with your project.
The configuration file is where you describe Environments, Components, and other metadata stack uses to operate.

First, each stack needs to have a name, as any number of "stacks" can be present on a given system.

    stack:
        name: stack-name

Next, you must define the environments that your project will deploy to. Environment configuration is a list of EnvironmentDescriptions
that map an environment name to a set of Activation conditions. Activation conditions can be environment variables, kubernetes contexts, or user confirmations.
For example, the following defines a local environment that is active if the kubernetes context is "docker-desktop" 

    environments:
      - name: local
        activation:
          context: docker-desktop

Finally, Components should be defined in the configuration file as a list of ComponentDescription objects. 
Components are logical groupings of kubernetes objects that may be defined by any number of containers, and at least one kubernetes manifest.

    components:
      - name: config
        requiredVariables:
          - ENV
        manifests:
          - ./deployments/config.yaml
      - name: app                                       
        requiredVariables:
          - PWD
          - HOME
        exposable: true
        containers:
          - dockerfile: ./containers/app/Dockerfile
            context: ./containers/app
            image: stack-app
        manifests:
          - ./deployments/app.yaml
          
Each component must be named, and should define a list of kubernetes manifests that make up the component.
Component's can also define a list of containers that the constituent manifests may depend on. These configurations
allow for command shorthands like `stack up app` and `stack build app` that will operate on all manifests, or containers respectively.

### Running

The Stack CLI requires a running kubernetes cluster to perform most commands. Locally, this will usually be Docker-Desktop, or Minikube.
See install instructions for [Docker Desktop](https://docs.docker.com/docker-for-mac/#kubernetes#kubernetes) and 
[Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).

Many stack commands refer back to the project configuration file for information on how to execute. To save typing, Stack 
will search the present working directory for a configuration file named `./stack-local.yaml`. 

⚠ **Ensure you are in a configured directory, or have explicitly provided a path to a project configuration file** 

In a directory with project components defined and configured, you may run `stack build` then `stack up` to bring up the
latest version of the stack.ommands.  

The stack CLI assumes the present working directory is the root project directory, and that a configuration file 
exists. Alternately, you can provide the desired root directory (with configuration file) by setting the `project_directory` flag on the root stack command.


In a directory with project components defined and configured, you may run `stack build` then `stack up` to bring up the
latest version of the stack.