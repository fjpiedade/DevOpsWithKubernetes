# Todo App – Exercise 1.13

This application is part of the DevOps with Kubernetes 2026 course.

## Overview

The Todo App is a Go application that displays a random image retrieved from [Lorem Picsum](https://picsum.photos/) and provides the initial user interface for a Todo application.

The application implements an image caching mechanism and uses persistent storage to preserve the downloaded image across container and Kubernetes Pod restarts.

The frontend is separated from the Go backend, with HTML and CSS files embedded into the Go executable using the `embed` package.

## Features

### Todo application interface

- Input field for entering new todos.
- Maximum input length of 140 characters.
- Send button (not connected to a backend yet).
- Todo list displaying predefined hardcoded tasks.
- Responsive layout.

### Image caching and persistence

- Downloads a random image from Lorem Picsum.
- Caches the downloaded image for 10 minutes.
- After the cache expires, serves the previous image one more time before downloading a new one on the next request.
- Serves the cached image if a new download fails and an existing image is available.
- Stores the image and its timestamp in a configurable directory.
- Supports persistent storage through Docker volumes and Kubernetes PersistentVolumes and PersistentVolumeClaims.

## Project structure

```text
todo_app/
├── Dockerfile
├── go.mod
├── main.go
├── static/
│   ├── index.html
│   └── styles.css
├── image/
├── manifests/
│   ├── deployment.yaml
│   ├── ingress.yaml
│   └── service.yaml
├── persistent-volume/
│   ├── persistentvolume.yaml
│   └── persistentvolumeclaim.yaml
└── README.md
```

## Frontend

The frontend is maintained separately from the Go application:

- `static/index.html`: application structure and hardcoded todo list.
- `static/styles.css`: application styling and responsive layout.

The Go application uses the `embed` package to include the frontend files in the compiled executable.

The todo input uses the HTML `maxlength="140"` attribute to limit the input to 140 characters.

The Send button is part of the interface but does not submit or persist todos at this stage.

## Image caching

The application uses the following files:

- `image.jpg`: downloaded image.
- `image.jpg.timestamp`: timestamp of the last successful download.
- `image.jpg.grace`: records whether the previous image has already been served once after cache expiration.

The image file location can be configured using the `IMAGE_FILE` environment variable.

Default value:

```text
/data/image.jpg
```

## Running locally

Start the application:

```bash
IMAGE_FILE=image/image.jpg PORT=9901 go run .
```

Open:

```text
http://localhost:9901
```

## Building the Docker image

Build the application image:

```bash
docker build -t fjpiedade/todo-app:1.4 .
```

## Running with Docker

Create a persistent volume:

```bash
docker volume create todo-app-data
```

Run the application:

```bash
docker run -d \
  --name todo-app \
  -p 9901:9901 \
  -v todo-app-data:/data \
  fjpiedade/todo-app:1.4
```

The Docker volume preserves the downloaded image when the container is removed and recreated.

Open:

```text
http://localhost:9901
```

## Docker Hub

The application image is available at:

[https://hub.docker.com/r/fjpiedade/todo-app](https://hub.docker.com/r/fjpiedade/todo-app)

Image tag:

```text
fjpiedade/todo-app:1.4
```

## Kubernetes deployment

The Kubernetes configuration is located in:

```text
persistent-volume/
manifests/
```

The application uses:

- A PersistentVolume (PV).
- A PersistentVolumeClaim (PVC).
- A Deployment.
- A ClusterIP Service.
- An Ingress managed by Traefik.

Apply the storage resources:

```bash
kubectl apply -f persistent-volume/
```

Apply the application resources:

```bash
kubectl apply -f manifests/
```

Check the deployment:

```bash
kubectl get pods
kubectl get pv
kubectl get pvc
kubectl get services
kubectl get ingress
```

## Persistence verification

Check the stored image:

```bash
kubectl exec deployment/todo-app -- sha256sum /data/image.jpg
```

Delete the application Pod:

```bash
kubectl delete pod -l app=todo-app
```

Wait for the replacement Pod:

```bash
kubectl get pods
```

Check the image hash again:

```bash
kubectl exec deployment/todo-app -- sha256sum /data/image.jpg
```

The hash should remain unchanged after Pod recreation, provided the cache has not expired and no new image download has occurred.

## Environment variables

| Variable     | Default           | Description         |
| ------------ | ----------------- | ------------------- |
| `PORT`       | `9901`            | HTTP server port    |
| `IMAGE_FILE` | `/data/image.jpg` | Image file location |

## Exercise 1.13 – Implementation summary

This exercise introduces the initial Todo application interface:

- Added a todo input field with a 140-character limit.
- Added a Send button.
- Added a list of predefined hardcoded todos.
- Separated the frontend HTML and CSS from the Go backend.
- Embedded frontend assets into the Go executable using `embed`.
- Preserved the existing image caching and persistent storage functionality.

The Send button does not create or save todos yet. Todo creation and persistence can be implemented in a later exercise.
