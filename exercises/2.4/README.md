# Todo App — Exercise 2.4

## Overview

This exercise continues the Todo application developed in Exercise 2.2.

The main objective is to create a dedicated Kubernetes namespace called `project` and deploy all project-related Kubernetes resources in that namespace.

The `project` namespace will also be used for future exercises related to the project. The `exercises` namespace remains dedicated to the regular course exercises.

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

## Kubernetes Namespace

A dedicated namespace called `project` is created for the project.

The namespace is defined in `namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: project
```

All namespaced application resources declare:

```yaml
metadata:
  namespace: project
```

This applies to:

- `todo-app` Deployment
- `todo-backend` Deployment
- `todo-app-svc` Service
- `todo-backend-svc` Service
- `todo-app-ingress` Ingress
- `persisting-data-pvc` PersistentVolumeClaim

**PersistentVolume note:** A PersistentVolume (PV) is a cluster-scoped resource and does not belong to a namespace. The PersistentVolumeClaim (PVC) is namespaced and belongs to `project`.

## Kubernetes resources

- `project` Namespace
- `todo-app` Deployment
- `todo-backend` Deployment
- `todo-app-svc` ClusterIP Service
- `todo-backend-svc` ClusterIP Service
- `todo-app-ingress` Ingress
- `persisting-data-pv` PersistentVolume
- `persisting-data-pvc` PersistentVolumeClaim

## Deployment

From the `exercises/2.4/` directory:

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

### 3. Deploy the application

```bash
kubectl apply -f manifests/
```

### 4. Check the resources

```bash
kubectl get all -n project
kubectl get ingress -n project
kubectl get pvc -n project
kubectl get pv
```

Both application Deployments should become ready.

## Inter-pod communication

The frontend connects to the backend through the Kubernetes Service DNS name:

```text
http://todo-backend-svc:80
```

Both Services and Deployments run in the same namespace, so Kubernetes DNS resolves the Service name directly.

The browser sends API requests to `/todos`. The `todo-app` forwards these requests to the backend using the `TODO_BACKEND_URL` environment variable.

## Testing

Open the application through the configured Ingress:

```text
http://localhost:8081
```

Retrieve todos:

```bash
curl http://localhost:8081/todos
```

Create a todo:

```bash
curl -X POST http://localhost:8081/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"run using kubernetes"}'
```

Check the application logs:

```bash
kubectl logs deployment/todo-app -n project
kubectl logs deployment/todo-backend -n project
```

## Image caching and persistence

The application downloads a random image from [Lorem Picsum](https://picsum.photos/) and stores it in persistent storage.

The image file location is configurable through the `IMAGE_FILE` environment variable.

Default value:

```text
/data/image.jpg
```

The PersistentVolume and PersistentVolumeClaim provide storage that survives application Pod restarts.

The Todo backend stores todos in memory. Todos are lost when the backend process restarts; persistent storage in this exercise is used for the cached image, not for todos.

## Project structure

```text
exercises/2.4/
├── manifests/
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

## Key concepts

- Kubernetes Namespaces
- Resource isolation and organization
- Cluster-scoped and namespaced resources
- Deployments and Services
- Kubernetes internal DNS
- Ingress
- PersistentVolumes and PersistentVolumeClaims
- Application deployment in a dedicated namespace

## Summary

Exercise 2.4 introduces a dedicated Kubernetes namespace for the Todo project.

All project-related namespaced resources are deployed in `project`, separating them from the regular course exercises.

The `project` namespace will be reused for future project-related exercises.
