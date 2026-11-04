# Todo App – Exercise 1.12

This application is part of the DevOps with Kubernetes course.

## Overview

The Todo App is a Go application that displays a random image retrieved from [Lorem Picsum](https://picsum.photos/).

The application implements an image caching mechanism to reduce external API requests and uses persistent storage to preserve the downloaded image across container and Pod restarts.

## Features

- Downloads a random image from Lorem Picsum.
- Caches the downloaded image for 10 minutes.
- After the cache expires, serves the previous image one more time before downloading a new one on the next request.
- Serves the cached image if a new download fails and an existing image is available.
- Stores the image and its timestamp in a configurable directory.
- Supports persistent storage through Docker volumes and Kubernetes PersistentVolumes and PersistentVolumeClaims.

## Image caching

The application uses the following files:

- `image.jpg`: downloaded image.
- `image.jpg.timestamp`: timestamp of the last successful download.
- `image.jpg.grace`: records whether the previous image has already been served once after cache expiration.

The image file location can be configured using the `IMAGE_FILE` environment variable. Its default value is `/data/image.jpg`.

## Running locally

```bash
IMAGE_FILE=image/image.jpg PORT=9901 go run .
```

Open:

```text
http://localhost:9901
```

## Building the Docker image

```bash
docker build -t fjpiedade/todo-app:1.3 .
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
  fjpiedade/todo-app:1.3
```

The Docker volume preserves the downloaded image when the container is removed and recreated.

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
| `PORT`       | `8080`            | HTTP server port    |
| `IMAGE_FILE` | `/data/image.jpg` | Image file location |

## Docker Hub

[https://hub.docker.com/r/fjpiedade/todo-app](https://hub.docker.com/r/fjpiedade/todo-app)
