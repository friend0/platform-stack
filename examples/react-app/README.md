# React App
The react-app stack example showcases an extremely common application pattern - a web frontend served by nginx, 
communicating with a separate backend server. The app itself is of bare-bones functionality - it simply
renders the default CRA home page, and makes a call to the backend API after mounting the app.


When the stack is brought up, the web frontend is served by Nginx, which sits behind a LoadBalancer service available at
localhost:31000. On start, the frontend fetches data from a golang backend server running in the cluster. Nginx
is configured to proxy to the backend, and the react development server is configured to proxy when running the command
locally over telepresence.

The configuration file for this example contains many helpful comments that walk through sections of the configuration.

## Running This Example

First, build all dependent containers:

    $ stack build all
    Building all containers for component `frontend`:
    Building image `react-frontend`:
    .
    .
    .
    Successfully built e3a0ef40978b
    Successfully tagged react-frontend:latest
    
    Building all containers for component `backend`:
    Building image `react-backend`:
    Sending build context to Docker daemon   12.8kB
    .
    .
    .
    Successfully built 8cd9ace54624
    Successfully tagged react-backend:latest
    
Next, bring up the stack:

    $ stack up
    Bringing up config
    configmap/react-app-env created
    Bringing up frontend
    service/frontend created
    deployment.apps/frontend created
    Bringing up backend
    service/backend created
    deployment.apps/backend created
    
Get pods running in the stack:

    $ stack pods
    NAME                                            READY   STATUS          RESTARTS        AGE             IP              NODE                    NOMINATED       READINESS       IMAGES
    backend-6c8668857d-q7gbl                        1/1     Running         0               8m39s           10.1.14.142     docker-desktop          <none>          <none>          [react-backend:latest]
    frontend-5fc5f74467-bzzrl                       1/1     Running         0               8m40s           10.1.14.141     docker-desktop          <none>          <none>          [react-frontend:latest]
    
Get stack health:

    $ stack health
    All pods are healthy
    ✔️  backend-6c8668857d-q7gbl in namespace `default` is healthy
    ✔️  frontend-5fc5f74467-bzzrl in namespace `default` is healthy

Visit the frontend page at `localhost:31000`. Opening your browser's console will show that the app has 
retrieved an array of todo items from the backend.
   
Get backend logs:

    $ stack logs backend

Get frontend logs:

    $ stack logs frontend

Expose a deployment by component name:

    $ stack enter backend
    
Enter a running container:

    $ stack enter backend
