# Todo App — Exercise 2.6

## Overview

This exercise continues the Todo application developed in previous exercises.

The main objective is to ensure that the application has no hardcoded infrastructure configuration, such as ports, service URLs, or runtime settings.

All configurable values are passed to the application containers through environment variables defined in Kubernetes ConfigMaps.

The application continues to run in the dedicated `project` namespace, keeping project-related resources separate from regular course exercises in the `exercises` namespace.

## Architecture

The project consists of two Go services:

- **todo-app**: serves the web interface, static assets and cached random image. It also acts as a reverse proxy for Todo API requests.
- **todo-backend**: provides a REST API to retrieve and create todos.

The frontend communicates with the backend through `/todos`. The backend stores todos in memory.

The application also uses persistent storage for the cached image.

```text
Browser
   |
   v
Ingress (Traefik)
   |
   v
todo-app-svc
   |
   v
todo-app (Go)
   |
   | Reverse proxy: /todos
   v
todo-backend-svc
   |
   v
todo-backend (Go)
   |
   v
In-memory todos
```

Both Ingress paths, `/` and `/todos`, are routed to `todo-app-svc`. The frontend handles API forwarding to the backend using `TODO_BACKEND_URL`.

## Kubernetes Namespace

All namespaced application resources are deployed in the `project` namespace.

The namespace is defined in `namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: project
```

The `project` namespace is separate from the `exercises` namespace.

A PersistentVolume (PV) is cluster-scoped and does not belong to a namespace. The PersistentVolumeClaim (PVC) belongs to `project`.

## Kubernetes Resources

- `project` Namespace
- `todo-app` Deployment
- `todo-backend` Deployment
- `todo-app-svc` ClusterIP Service
- `todo-backend-svc` ClusterIP Service
- `todo-app-ingress` Ingress
- `todo-app-config` ConfigMap
- `todo-backend-config` ConfigMap
- `persisting-data-pv` PersistentVolume
- `persisting-data-pvc` PersistentVolumeClaim

## Configuration Management

The application does not rely on hardcoded infrastructure configuration.

Runtime settings are provided through environment variables defined in Kubernetes ConfigMaps.

### todo-app-config

The `todo-app-config` ConfigMap contains:

| Variable               | Description                    | Value                        |
| ---------------------- | ------------------------------ | ---------------------------- |
| `PORT`                 | Application listening port     | `9901`                       |
| `IMAGE_FILE`           | Cached image file location     | `/data/image.jpg`            |
| `TODO_BACKEND_URL`     | Backend Kubernetes Service URL | `http://todo-backend-svc:80` |
| `IMAGE_SOURCE_URL`     | External image source          | `https://picsum.photos/1200` |
| `IMAGE_CACHE_DURATION` | Image cache duration           | `10m`                        |
| `IMAGE_HTTP_TIMEOUT`   | HTTP client timeout            | `30s`                        |

### todo-backend-config

The `todo-backend-config` ConfigMap contains:

| Variable | Description            | Value  |
| -------- | ---------------------- | ------ |
| `PORT`   | Backend listening port | `8080` |

Both Deployments load their respective ConfigMaps using `envFrom`.

The Go applications require their configuration environment variables to be present. Missing or invalid required configuration causes the application to fail rather than silently using hardcoded defaults.

## Deployment

From the `exercises/2.6/` directory:

### 1. Create the namespace

```bash
kubectl apply -f namespace.yaml
```

### 2. Apply persistent storage

```bash
kubectl apply -f persistent-volume/
```

Check the storage resources:

```bash
kubectl get pv
kubectl get pvc -n project
```

The PersistentVolumeClaim should reach the `Bound` status.

### 3. Apply ConfigMaps

```bash
kubectl apply -f manifests/configmap.yaml
```

### 4. Build and publish the Docker images

Build the backend image:

```bash
docker build -t fjpiedade/todo-backend:2.6 ./todo-backend
```

Build the frontend image:

```bash
docker build -t fjpiedade/todo-app:2.6 ./todo-app
```

Push both images to Docker Hub:

```bash
docker push fjpiedade/todo-backend:2.6
docker push fjpiedade/todo-app:2.6
```

### 5. Deploy the backend and frontend

```bash
kubectl apply -f manifests/todo-backend-deployment.yaml
kubectl apply -f manifests/todo-app-deployment.yaml
```

### 6. Apply Services and Ingress

```bash
kubectl apply -f manifests/todo-backend-service.yaml
kubectl apply -f manifests/todo-app-service.yaml
kubectl apply -f manifests/ingress.yaml
```

### 7. Check the resources

```bash
kubectl get pods -n project
kubectl get deployments -n project
kubectl get services -n project
kubectl get configmaps -n project
kubectl get ingress -n project
kubectl get pvc -n project
kubectl get pv
```

Both application Deployments should become ready, and the PVC should be `Bound`.

## Inter-pod Communication

The frontend connects to the backend through the Kubernetes Service DNS name:

```text
http://todo-backend-svc:80
```

Both Services and Deployments run in the same namespace, so Kubernetes DNS resolves the Service name directly.

The browser sends API requests to `/todos`. The `todo-app` forwards these requests to the backend using the `TODO_BACKEND_URL` environment variable.

The backend listens on the port defined by its `PORT` environment variable.

## Testing

The application can be tested through the configured Ingress or directly through its Kubernetes Service.

### Test through the Service

Start a port-forward:

```bash
kubectl port-forward -n project svc/todo-app-svc 8090:80
```

In another terminal, retrieve the application HTML:

```bash
curl http://localhost:8090/
```

Retrieve todos:

```bash
curl http://localhost:8090/todos
```

Create a todo:

```bash
curl -X POST http://localhost:8090/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"Exercise 2.6 configuration test"}'
```

Retrieve the todos again:

```bash
curl http://localhost:8090/todos
```

Check the image endpoint:

```bash
curl -I http://localhost:8090/image
```

The image endpoint should return HTTP `200 OK` with `Content-Type: image/jpeg`.

### Check application logs

```bash
kubectl logs deployment/todo-app -n project
kubectl logs deployment/todo-backend -n project
```

## Image Caching and Persistence

The application downloads a random image from [Lorem Picsum](https://picsum.photos/) and stores it in persistent storage.

The following settings are configurable through environment variables:

- `IMAGE_FILE`: image file location.
- `IMAGE_SOURCE_URL`: external image source.
- `IMAGE_CACHE_DURATION`: cache duration.
- `IMAGE_HTTP_TIMEOUT`: HTTP client timeout.

The PersistentVolume and PersistentVolumeClaim provide storage that survives application Pod restarts.

The Todo backend stores todos in memory. Todos are lost when the backend process restarts. Persistent storage in this exercise is used for the cached image, not for todos.

## Project Structure

```text
exercises/2.6/
├── manifests/
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── todo-app-deployment.yaml
│   ├── todo-app-service.yaml
│   ├── todo-backend-deployment.yaml
│   └── todo-backend-service.yaml
├── persistent-volume/
│   ├── persistentvolume.yaml
│   └── persistentvolumeclaim.yaml
├── todo-app/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── image/
│   └── static/
├── todo-backend/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── namespace.yaml
└── README.md
```

## Cleanup

To remove the application resources:

```bash
kubectl delete -f manifests/
kubectl delete -f persistent-volume/
```

To remove the namespace after removing its namespaced resources:

```bash
kubectl delete namespace project
```

Only remove the namespace when it is no longer needed by the project.

## Key Concepts

- Kubernetes ConfigMaps
- Environment variables in containers
- Externalized application configuration
- Docker image builds and publishing
- Kubernetes Namespaces
- Deployments and Services
- Kubernetes internal DNS
- Ingress
- PersistentVolumes and PersistentVolumeClaims
- Application configuration without hardcoded infrastructure settings

## Summary

Exercise 2.6 focuses on externalizing application configuration.

Ports, service URLs, image source settings, cache duration and HTTP timeout are passed to the application containers through environment variables defined in ConfigMaps.

The Todo application and backend are deployed independently in the `project` namespace, communicating through Kubernetes Services.

Persistent storage is used for the cached image, while todos remain stored in backend memory.
