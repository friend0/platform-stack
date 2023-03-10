# [Stack](stack)

Stack is a command-line workflow for mapping your application's deployment to a Kubernetes cluster. 
Stack bootstraps common deployment and debug workflows onto your applications by reading a configuration file specifying the environments, containers, and manifests that make it up. 

Example configuration:
```
    apiVersion: stack/v1alpha1
    stack:                                  
        name: aSimpleApp
    environments:
      - name: local
        activation:
          context: docker-desktop || minikube          
    components:
      - name: config
        requiredVariables:
          - ENV
        manifests:
          - ./deployments/config.yaml
      - name: app                                       
        containers:
          - dockerfile: ./containers/app/Dockerfile
            context: ./containers/app
            image: stack-app
        manifests:
          - ./deployments/app.yaml
```
Then, in the directory of this config (or by passing a path to a config file as an option) run the follwing:

`$ stack up`

This will nominally show all the components defined for this stack have been brought up successfully. 
Alternatively, to bring up only a subset of the components, run:

`$ stack up config app`

In the above, only the components `config` and `app` are brought up.
# [Prerequisites](prereqs)

The Stack CLI requires a running kubernetes cluster to perform most commands. Locally, this will usually be Docker-Desktop, or Minikube.
[Docker Desktop](https://docs.docker.com/docker-for-mac/install/)
[Docker Desktop Kubernetes](https://docs.docker.com/docker-for-mac/#kubernetes#kubernetes)

## [Step 1: Install the Stack CLI](install)

- Option 1: Install `jq` with `brew install jq`, then run the install script `install.sh`.
You will need to export a github personal access token as GIT_TOKEN `export GIT_TOKEN=<GENERATED_TOKEN_HERE>` [see here](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line).
- Option 2: Navigate to https://github.com/altiscope/platform-stack/releases and take the latest.
Next, put the appropriate binary onto your path renamed as `stack`
- Option 3: Build from source `go build -o /usr/local/bin/stack -v ../platform-stack/main.go`

Once stack is available, stack CLI dependencies can be installed by running 
```stack install```.
  
This installation will install xcode, and the supported kubectl, and kubetpl versions. 

## [Step 2: Define Kubernetes Manifests and a Stack Configuration file](config)

Stack is a tool for operating applications built on Kubernetes. In order for the Stack CLI to run your applicaion, you need:
    - A Stack configuration file
    - A set of Kubernetes YAML manifests
    - Dockerfile container definitions (if applicable)
    
### [The Stack Configuration File](stack-config) 
                
The configuration file is where you describe the Environments, and Components needed to run your application.
By default, this file should be named `.stack-local.yaml`, and should be included at the base directory of the project.
The following example shows configuration for a simple app with configuration.

    apiVersion: stack/v1alpha1
    stack:                                  
        name: example-stack
    environments:
      - name: local
        activation:
          context: docker-desktop
      - name: staging
        activation: 
          context: platform-stg-hjkabsy12
      - name: production
        activation:
          context: platform-prod-asku7112a
          confirmWithUser: true               
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
          - dockerfile: ./containers/app/Dockerfile-dev
            context: ./containers/app
            image: stack-app-live
            environments:
              - local
        manifests:
          - ./deployments/app.yaml

There are currently four main components of a Stack configuration file:
- ApiVersion: Used to maintain compatibility of configs with the latest Stack CLI as new features are added
- Stack: Metadata about the stack
- Environments: Description of the environments the stack deploys in
- Components: Description of the k8s manifests, env, etc. related to deploying a particular component

#### [Stack](stack-description)

    type Stack {
        Name string                             # The name of the stack - used to scope certain commands to Kubernetes labels (see above)
    }

At the top you'll notice a stack > name definition - this enables us to proerly scope Stack commands to the current project.

#### [Environments](environment-descriptions)

    type Environment {
        Name       string                       # The name of the Environment
        Activation ActivationDescription        # A description of conditiond under which this environment will be active
    }

    type Activation {
        ConfirmWithUser bool                    # Constructive and destructive operations require user confirmation
        Env             string                  # Environment key:value pair that must be set as `VARNAME=VALUE`
        Context         string                  # Kubernetes context that will activate this environment
    }

The `environments` section defines the Kubernetes contexts that correspond to the various environments you application can run against.
In the example above, the environments needed are local, staging, and production.
These environment configurations tell Stack which Kubernetes contexts to use for Stack operations. Conversely, 
the current Kubernetes context tells the Stack which contexts to use when running Stack commands.

In the `activation` section for each environment, you must specify the name of the kubernetes context you'd like to use 
as an activation condition for the given environment. For example, the environment "local" above will be active if the Kubernetes context is "docker-desktop".
The environment "production" will be active if the current context is "platform-prod-asku7112a", and constructive or destructive
Stack commands like `up` and `down` will only run after confirming with the user.
     
     
#### [Components](component-description)

    Component {
        Name              string                 # The name we'll use to refer to the component
        Environments      []string               # The environment(s) for which this component should be applied. 
        RequiredVariables []string               # A list of environment variables that mus tbe present on the system at runtime
        Exposable         bool                   # Should this component be exposable via kubectl port-forward?
        Containers        []Container            # A list of dependent container descriptions
        Manifests         []string               # A list of paths to kubernetes manifests that make up this component
    }

    type Container {
        Dockerfile   string                        # Relative path to Dockerfile
        Context      string                        # Relative path of context to build Dockerfile
        Image        string                        # The name of the image to be built from container
        Environments []string                      # The environment(s) for which this image should be built. 
                                                   # Leave blank to build for all environments
    } 
    
Components are logical groupings of kubernetes objects. Each component requires at least one kubernetes manifest, 
and any number of containers.

Each component requires a name. It must also define a list of Kubernetes manifests that make up the components.
A list of containers that the Component needs to run can also be included, allowing us to build containers the the Kubernetes 
manifests depend on before we try to bring up the cluster. 

The logical groupings that Components provide allow us to use easy shorthands like `stack up app` and `stack build app` 
that will operate on all manifests, or containers defined by the component named `app`.

### [Kubernetes Manifest Label Requirements](kubernetes-config) 

In order for Stack to properly scope certain commands to objects owned by a particular stack, **we require that 
kubernetes objects be defined with two required labels in their metadata**:

      labels:
        stack: example-stack        # Must correspond with the name of the current Stack
        app: backend                # should correspond to the Component it belongs to

Currently there is no validation of this step, so make sure to double check your Kubernetes YAML definitions! You can
check manifest in the examples directory to see this in practice.
        
        


## Step 3: Build Dependent Images and Run the Stack

??? **Ensure you are in a configured directory, or have explicitly provided a path to a stack configuration file** 

The stack CLI assumes the present working directory is the root project directory, and that a configuration file 
exists. Alternately, you can provide the desired root directory (with configuration file) by setting the `stack_directory` flag on the root stack command.

Build all dependent containers for the stack by running:

    stack build all

To build containers piecewise, run:

    stack build <COMPONENT> [CONTAINER]

Run the help command for more detailed options.

    stack help build
    
Next, bring up the entire stack with:

    stack up

## [Step 4: Manage the App](manage)

### Expose
If your app is running behind certain Kubernetes Services, you may need to port forward traffic from your local machine to the cluster.
This can be done by running:

    stack expose <component> <local port> <remote port>

See `stack help expose` for more details. This might not be necessary for types like ingress controllers and load balancers.

### Logs
You can get logs for a given deployment with: 

    stack logs [DEPLOYMENT_NAME]

### Health
You may check the health of the current cluster by running:
    
    stack health

Note that new deployments can take a few moments to become healthy.  

### Pods
Get running pods for the current Stack
 
    stack pods

### Deploy to Target Environments
Deploy to a remote environment by configuring your KUBECONFIG and associating Kubernetes contexts with environments
defined in your stack configuration file. 

Check your current environment:

    stack environment
    
Change your environment:

    stack environment staging
    
All operations will now be scoped to the current environment and context.

### Add and Remove Stack Secrets
As a convenience, the Stack CLI provides the secrets command for creating stock Kubenretes secret resources like those
used for imagePullSecrets and so on.

List the currently available secrets for the stack:

    stack secrets

Create a registry secret for authenticating with private container registries:

    stack secrets registry

Remove all secrets:

    stack secrets delete

Remove a specific secret:

    stack secrets delete registry       # deletes only the registry secret created above

### Fetch Secrets from GCP Secret Manager
Stack CLI provides workflow to fetch application runtime secrets from GCP Secret Manager (GSM).

    stack secrets fetch [-e <env>] [-p <gcp-project-id>] [-i <input-file-directory>] [-o <output-file-directory>] [flags]
    
Run `stack secrets fetch -h` to find details about the parameters.

## [Examples](examples)

If you would like to use the Stack CLI without first configuring your own project, you can navigate to the examples 
directory to get a feel for how to setup projects, and how stack works.

- [Basic Application](./examples/basic/README.md): A lightweight dummy application for testing out stack commands 
- [Nginx/React/Go Web Application](./examples/react-app/README.md): A prototypical web application with a backend, and
frontend serving up compiled assets. This example uses CRA for a simple web frontend. The binary is built and served up by Nginx at runtime,
and calls out to a Golang backend.

# Develop and Release
Releasing a new version of the `stack` binary requires a `git tag` which is automatically generated via [`semantics`](github.com/stevenmatthewt/semantics) and created via [`ghr`](github.com/tcnksm/ghr) from the [commit messages](https://github.com/stevenmatthewt/semantics#how-it-works) when a PR is merged to `master`. The release process is opinionated and requires planning when you [start developing a feature](https://github.com/stevenmatthewt/semantics#faq). Please follow these steps:
1. Checkout your feature branch `<your_github_username>/<Jira_ID>_feature_name` and open a PR against `master`
2. Commit changes to your branch as usual with typical commit messages
3. Decorate __only one__ of the `commit messages` in your PR with one of the prefixes `major:, minor:, patch:` which will be automatically used to create a release tag
  - __CAVEATS__:
  - __DO NOT decorate__ more than one commit messages with above prefixes.  Having more than one commits with those prefix will bump release tag version for each of those commits.
  - __DO NOT force push__ 
  - __DO NOT create a git tag manually__
4. Merge your PR with thumbs from PR reviewer
  - If all the previous steps are done correctly, then a new release will be created with new binaries. 

__References__:
- https://github.com/stevenmatthewt/semantics#how-it-works
- https://github.com/stevenmatthewt/semantics#faq
