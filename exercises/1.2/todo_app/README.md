# Todo Application

A simple HTTP server written in Go, created for Exercise 1.2 of the [DevOps with Kubernetes 2026 course](https://courses.mooc.fi/org/uh-cs/courses/devops-with-kubernetes-2026).

## Application

The server:

- Starts an HTTP server.
- Uses the `PORT` environment variable to select the listening port.
- Defaults to port `9901` when `PORT` is not set.
- Returns `To-do Application with Go` for HTTP requests to `/`.

At startup, the server logs a message such as:

```text
Server started in port 9000
```

## Exercise 1.6 - The project, step 4

### 1. Create the k3d cluster

```bash
k3d cluster create --port 8082:30080@agent:0 -p 8081:80@loadbalancer --agents 2
```

### 2. Deploy the application

```bash
kubectl create deployment todo-app --image=fjpiedade/todo-app:1.0
kubectl rollout status deployment/todo-app
```

### 3. Get the Service manifest

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes
```

**Note:** If you already cloned the repository, skip this step and run the following commands from the repository root.

### 4. Apply the NodePort Service

```bash
kubectl apply -f exercises/1.2/todo_app/manifests/service.yaml
```

### 5. Verify the deployment and Service

```bash
kubectl get deployments
kubectl get pods
kubectl get services
```

### 6. Test the application

```bash
curl http://localhost:8082/
```

Expected response:

```text
To-do Application with Go
```

## Docker

Build the image:

```bash
docker build -t fjpiedade/todo-app:1.0 .
```

Run the container:

```bash
docker run --rm -e PORT=9000 -p 9000:9000 fjpiedade/todo-app:1.0
```

The image is published on [Docker Hub](https://hub.docker.com/r/fjpiedade/todo-app).

## Kubernetes

The application is deployed using imperative `kubectl` commands, without a deployment manifest.

```bash
kubectl create deployment todo-app --image=fjpiedade/todo-app:1.0
kubectl set env deployment/todo-app PORT=9000
kubectl rollout status deployment/todo-app
kubectl logs deployment/todo-app
```

The deployment was verified in Kubernetes, and the logs confirmed that the server started on port `9000`.

External access to the application is not configured in this exercise.

## Kubernetes with Declarative approach using manifests

The application is deployed to Kubernetes using a declarative Deployment manifest located at `manifests/deployment.yaml`.

Apply the manifest from the repository root:

```bash
kubectl apply -f exercises/1.2/todo_app/manifests/deployment.yaml
```

Check the Deployment and Pods:

```bash
kubectl rollout status deployment/todo-app
kubectl get deployments
kubectl get pods
```

Check the application logs:

```bash
kubectl logs deployment/todo-app
```

The application listens on port `9901` by default. External access is not configured in this exercise.

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- kubectl
