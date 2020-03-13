
# ☰ Stack

Stack is a tool for defining and running Kubernetes applications. 
With Stack, you use a configuration file to define the services that make up your application. 
Then, with a few simple commands, you build and deploy all the services from your configuration. 

Stack is a generalized CLI for seamless test, development, and debugging across environments.
Stack aims to give developers a powerful set of tools for developing and maintaining services across environment, 
helping to minimize differences between development and production.


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
By default, this file should be named `.stack-local.yaml`, and should be included at the base directory of the project.

First, each stack needs to have a name, as any number of "stacks" can be present on a given system.

    stack:
        name: example-stack

Next, you must define the environments that your project will deploy to. Environment configuration is a list of EnvironmentDescriptions
that map an environment name to a set of Activation conditions. Activation conditions can be environment variables, kubernetes contexts, or user confirmations.
For example, the following defines a local environment that is active if the kubernetes context is "docker-desktop" and the
variable `ENV` is set to `local`.

    environments:
      - name: local
        activation:
          context: docker-desktop
          env: ENV=local            

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
Components can also define a list of containers that the constituent manifests may depend on. These configurations
allow for command shorthands like `stack up app` and `stack build app` that will operate on all manifests, or containers respectively.

## Examples

If you would like to use the Stack CLI without first configuring your own project, you can navigate to the examples 
directory to get a feel for how to setup projects, and how stack works.

- [Basic Application](./examples/basic/README.md): A lightweight dummy application for testing out stack commands 
- [Nginx/React/Go Web Application](./examples/react-app/README.md): A prototypical web application with a backend, and
frontend serving up compiled assets. This example uses CRA for a simple web frontend. The binary is built and served up by Nginx at runtime,
and calls out to a golang backend.

## Running Examples

The Stack CLI requires a running kubernetes cluster to perform most commands. Locally, this will usually be Docker-Desktop, or Minikube.
See install instructions for [Docker Desktop](https://docs.docker.com/docker-for-mac/#kubernetes#kubernetes) and 
[Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).

The stack CLI assumes the present working directory is the root project directory, and that a configuration file 
exists. Alternately, you can provide the desired root directory (with configuration file) by setting the `project_directory` flag on the root stack command.

⚠ **Ensure you are in a configured directory, or have explicitly provided a path to a stack configuration file** 
 
To run any of the example app, first check which environments are available by running `stack environment list`.
If an environment is active, it will show up green in the list. If none are active, you can run `stack environment local` - 
be sure to note any environment variables that need to be set, then confirm the environment by repeating the command above. 

Next, check which pods may already be running in the current environment by running `stack pods`.

In a directory with project components defined and configured, you may run `stack build` then `stack up` to bring up the
latest version of the stack. You can then run `stack pods` again to see

You may check the health of the current cluster by running `stack health` - tt can take a few moments for a new deployment
to come up.  

When your stack is healthy, you can start tailing logs with `stack logs -f [DEPLOYMENT_NAME]`.

You can enter a currently running container on a target pod with `stack enter [DEPLOYMENT_NAME]`.


In a directory with project components defined and configured, you may run `stack build` then `stack up` to bring up the
latest version of the stack.