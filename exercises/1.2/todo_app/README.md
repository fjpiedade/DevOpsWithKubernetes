# Todo App — Exercise 1.2

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

## Run locally

```bash
go run .
```

To select a different port:

```bash
PORT=9000 go run .
```

Test the HTTP response:

```bash
curl http://localhost:9000/
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

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- kubectl
