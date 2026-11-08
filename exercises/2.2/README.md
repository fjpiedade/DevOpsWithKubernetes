# Todo App – Exercise 2.2

This application is part of the DevOps with Kubernetes 2026 course.

## Overview

Exercise 2.2 extends the Todo application by separating the frontend application from the Todo backend.

The project consists of two Go services:

- **todo-app**: serves the web interface, static assets and cached random image. It also acts as a reverse proxy for Todo API requests.
- **todo-backend**: provides a REST API to retrieve and create todos.

The frontend communicates with the backend through `/todos`. The backend stores todos in memory; no database is used in this exercise.

The application also maintains the image caching mechanism introduced in previous exercises. The downloaded image is stored on persistent storage so that it can survive container and Kubernetes Pod restarts.

## Architecture

```text
Browser
   |
   | HTTP requests
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

The browser always sends API requests to `/todos`. The `todo-app` forwards these requests to the backend using the `TODO_BACKEND_URL` environment variable.

This approach avoids exposing the backend directly to the browser and avoids the need for CORS configuration.

The backend address depends on the environment:

| Environment | `TODO_BACKEND_URL`           |
| ----------- | ---------------------------- |
| Local       | `http://localhost:8080`      |
| Docker      | `http://todo-backend:8080`   |
| Kubernetes  | `http://todo-backend-svc:80` |

## Features

### Todo application

- Web interface for creating todos.
- Maximum todo text length of 140 characters.
- Dynamic todo list loaded from the backend.
- Todos can be created using the Send button.
- REST API communication using JavaScript `fetch()`.
- JSON responses.
- Backend request logging.
- Todos are stored in memory and are not persisted across backend restarts.

### Todo backend API

The backend exposes the following endpoints:

| Method | Endpoint | Description       | Response      |
| ------ | -------- | ----------------- | ------------- |
| GET    | `/todos` | Returns all todos | `200 OK`      |
| POST   | `/todos` | Creates a todo    | `201 Created` |

Example request:

```bash
curl -X POST http://localhost:9901/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"My first todo"}'
```

Example response:

```json
{
  "id": 1,
  "text": "My first todo"
}
```

Retrieve todos:

```bash
curl http://localhost:9901/todos
```

The backend validates the request body, rejects empty todo text and enforces the 140-character limit.

### Image caching and persistence

- Downloads a random image from [Lorem Picsum](https://picsum.photos/).
- Caches the downloaded image for 10 minutes.
- After cache expiration, serves the previous image one more time before attempting a new download on the next request.
- Serves the existing cached image if a new download fails.
- Stores the image and timestamp in a configurable directory.
- Supports persistent storage through Docker volumes and Kubernetes PersistentVolumes and PersistentVolumeClaims.

The image file location is configurable through the `IMAGE_FILE` environment variable.

Default value:

```text
/data/image.jpg
```

## Project structure

```text
exercises/2.2/
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
│   └── static/
│       ├── app.js
│       ├── index.html
│       └── styles.css
├── todo-backend/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── README.md
```

The frontend HTML, CSS and JavaScript files are embedded into the `todo-app` Go executable using the `embed` package.

## Running locally

The backend and frontend run as separate processes.

### 1. Start the backend

Open a terminal in `todo-backend/`:

```bash
PORT=8080 go run .
```

The backend listens on port `8080`.

### 2. Start the frontend

Open another terminal in `todo-app/`:

```bash
TODO_BACKEND_URL=http://localhost:8080 \
IMAGE_FILE=image/image.jpg \
PORT=9901 \
go run .
```

Open the application:

```text
http://localhost:9901
```

The frontend forwards `/todos` requests to the backend running on port `8080`.

## Building the Docker images

Build the backend image:

```bash
cd exercises/2.2/todo-backend
docker build -t fjpiedade/todo-backend:2.2 .
```

Build the frontend image:

```bash
cd ../todo-app
docker build -t fjpiedade/todo-app:2.2 .
```

## Running with Docker

Create a Docker network so that both containers can communicate:

```bash
docker network create todo-test-net
```

Create a persistent volume for the image:

```bash
docker volume create todo-app-data
```

### 1. Start the backend

```bash
docker run -d \
  --name todo-backend \
  --network todo-test-net \
  fjpiedade/todo-backend:2.2
```

### 2. Start the frontend

```bash
docker run -d \
  --name todo-app \
  --network todo-test-net \
  -p 9901:9901 \
  -e TODO_BACKEND_URL=http://todo-backend:8080 \
  -e IMAGE_FILE=/data/image.jpg \
  -e PORT=9901 \
  -v todo-app-data:/data \
  fjpiedade/todo-app:2.2
```

Open:

```text
http://localhost:9901
```

Test the API through the frontend:

```bash
curl http://localhost:9901/todos
```

Create a todo:

```bash
curl -X POST http://localhost:9901/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"Test Docker communication"}'
```

The Docker network allows the frontend container to resolve the backend using its container name.

## Docker Hub

The application images are available at:

- [todo-app](https://hub.docker.com/r/fjpiedade/todo-app)
- [todo-backend](https://hub.docker.com/r/fjpiedade/todo-backend)

Exercise 2.2 image tags:

```text
fjpiedade/todo-app:2.2
fjpiedade/todo-backend:2.2
```

## Kubernetes deployment

The Kubernetes configuration is located in:

```text
persistent-volume/
manifests/
```

The deployment uses:

- Two Deployments: `todo-app` and `todo-backend`.
- Two ClusterIP Services: `todo-app-svc` and `todo-backend-svc`.
- A Traefik Ingress for external access to the frontend.
- A PersistentVolume (PV).
- A PersistentVolumeClaim (PVC).

The frontend connects to the backend through the Kubernetes Service DNS name:

```text
http://todo-backend-svc:80
```

### 1. Apply persistent storage

From the `exercises/2.2/` directory:

```bash
kubectl apply -f persistent-volume/
```

Check the storage resources:

```bash
kubectl get pv,pvc
```

Both resources should reach the `Bound` status.

### 2. Apply application resources

```bash
kubectl apply -f manifests/
```

### 3. Check the deployment

```bash
kubectl get deployments,pods,services,ingress
```

Check the Services and their endpoints:

```bash
kubectl get endpoints
```

Both application Services should have endpoints when their Pods are ready.

### 4. Test the application

Open the application through the configured Ingress.

In the local Kubernetes environment used for this exercise:

```text
http://localhost:8081
```

Test the API through the frontend:

```bash
curl http://localhost:8081/todos
```

Create a todo:

```bash
curl -X POST http://localhost:8081/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"run using kubernetes"}'
```

Check the backend logs:

```bash
kubectl logs deployment/todo-backend
```

Check the frontend logs:

```bash
kubectl logs deployment/todo-app
```

## Persistence verification

The image is stored in `/data/image.jpg`, which is mounted from the PersistentVolumeClaim.

Check the stored image:

```bash
kubectl exec deployment/todo-app -- ls -lh /data/
```

Calculate its hash:

```bash
kubectl exec deployment/todo-app -- sha256sum /data/image.jpg
```

Delete the application Pod:

```bash
kubectl delete pod -l app=todo-app
```

Wait for Kubernetes to create a replacement Pod:

```bash
kubectl get pods -w
```

After the replacement Pod is ready, check the image hash again:

```bash
kubectl exec deployment/todo-app -- sha256sum /data/image.jpg
```

The hash should remain unchanged if the cached image has not been replaced.

## Environment variables

### todo-app

| Variable           | Default                 | Description                           |
| ------------------ | ----------------------- | ------------------------------------- |
| `PORT`             | `9901`                  | HTTP server port                      |
| `IMAGE_FILE`       | `/data/image.jpg`       | Image file location                   |
| `TODO_BACKEND_URL` | `http://localhost:8080` | Backend URL used by the reverse proxy |

### todo-backend

| Variable | Default | Description      |
| -------- | ------- | ---------------- |
| `PORT`   | `8080`  | HTTP server port |

## Exercise 2.2 – Implementation summary

This exercise extends the Todo application with a separate backend service.

- Created a Go backend service.
- Implemented `GET /todos` to retrieve todos.
- Implemented `POST /todos` to create todos.
- Added JSON request and response handling.
- Added input validation and a 140-character limit.
- Stored todos in memory using a synchronized data structure.
- Added request logging to both Go services.
- Updated the frontend to load and create todos using JavaScript `fetch()`.
- Added a reverse proxy in `todo-app` to forward `/todos` requests to the backend.
- Configured the backend URL through the `TODO_BACKEND_URL` environment variable.
- Created separate Docker images for the frontend and backend.
- Deployed both services to Kubernetes using Deployments and ClusterIP Services.
- Configured Traefik Ingress to expose the frontend.
- Continued using persistent storage for the cached image through a PV and PVC.

The Todo backend uses in-memory storage. Todos are therefore lost when the backend process restarts. Persistent storage in this exercise is used for the cached image, not for todos.
