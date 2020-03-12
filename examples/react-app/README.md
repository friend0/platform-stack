# React App
The react-app stack example showcases an extremely common application pattern - a web frontend served by nginx, 
communicating with a separate backend server. 

When the stack is brought up, the web frontend is served by Nginx, which sits behind a LoadBalancer service available at
localhost:3100. On start, the frontend fetches data from a golang backend server running in the cluster. Nginx
is configured to proxy to the backend, and the react development server is configured to proxy when running the command
locally over telepresence.

The configuration file for this example contains many helpful comments that walk through sections of the configuration.