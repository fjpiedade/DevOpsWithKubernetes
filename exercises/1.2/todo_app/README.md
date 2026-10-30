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

# Exercise 1.8 — The project, step 5

## Overview

The Todo application is now accessible externally through a Kubernetes Ingress instead of a NodePort Service.

Traefik routes HTTP requests to the Todo application.

## 1. Clone the repository

Clone the project:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes
```

If you already have the repository cloned, skip this step and run the following commands from the repository root.

## 2. Kubernetes resources

The application uses the following resources:

- **Deployment:** `todo-app`
- **Service:** `todo-app-svc` (`ClusterIP`)
- **Ingress:** `todo-app-ingress` (`traefik`)
- **Application port:** `9901`

The manifests are located in `exercises/1.2/todo_app/manifests/`:

- `deployment.yaml` — deploys the Todo application.
- `service.yaml` — exposes the application internally using ClusterIP.
- `ingress.yaml` — provides external HTTP access through Traefik.

## 3. Deploy the application

Apply the Kubernetes manifests:

```bash
kubectl apply -f exercises/1.2/todo_app/manifests/
```

Check the resources:

```bash
kubectl get deployments
kubectl get services
kubectl get ingress
kubectl get pods
```

## 4. Access the application

The application is accessible through the Ingress at:

```text
http://localhost:8081/
```

Test it with:

```bash
curl http://localhost:8081/
```

Expected response:

```text
To-do Application with Go
```

## Result

The Todo application is accessible through Traefik Ingress using a ClusterIP Service instead of NodePort.

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
