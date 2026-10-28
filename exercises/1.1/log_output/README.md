# 1.1 - Log Output

A Go application that generates a random UUID when it starts and prints it to the console every five seconds, together with the current UTC timestamp.

This exercise introduces basic application development, Docker containerization, publishing images to Docker Hub, and deploying an application to Kubernetes using k3d.

## Application behavior

- Generates a UUID when the application starts.
- Keeps the same UUID in memory for the lifetime of the process.
- Prints the current UTC timestamp and UUID every five seconds.
- Writes output to standard output so that it can be inspected using Docker logs or `kubectl logs`.

Example output:

```text
2026-10-28T00:52:46.877Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
2026-10-28T00:52:51.880Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
2026-10-28T00:52:56.886Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
```

The UUID remains the same while the timestamps change.

## Project structure

```text
log_output/
├── Dockerfile
├── go.mod
├── main.go
├── manifests/
│   └── deployment.yaml
└── README.md
```

## Run locally

Run the application using Go:

```bash
go run .
```

Stop the application with `Ctrl+C`.

## Build and run with Docker

Build the Docker image:

```bash
docker build -t log-output:1.0 .
```

Run the container:

```bash
docker run --rm --name log-output-test log-output:1.0
```

Stop the container with `Ctrl+C`.

## Docker Hub

The image is published on Docker Hub:

- [`fjpiedade/log-output:1.0`](https://hub.docker.com/r/fjpiedade/log-output)
- [`fjpiedade/log-output:latest`](https://hub.docker.com/r/fjpiedade/log-output)

Run the published image:

```bash
docker run --rm fjpiedade/log-output:1.0
```

Alternatively, use the `latest` tag or not tag:

```bash
docker run --rm fjpiedade/log-output:latest
```

```bash
docker run --rm fjpiedade/log-output
```

## Deploy to Kubernetes

This exercise uses Kubernetes through a k3d cluster.

Apply the Deployment manifest from this directory:

```bash
kubectl apply -f manifests/deployment.yaml
```

Wait for the Deployment to become available:

```bash
kubectl rollout status deployment/log-output
```

Check the Pod status:

```bash
kubectl get pods
```

View the application logs:

```bash
kubectl logs -f deployment/log-output
```

The logs should show the same UUID every five seconds, with a new UTC timestamp on each line.

Stop following the logs with `Ctrl+C`. This does not stop the application running in Kubernetes.

## Cleanup

Delete the Deployment and its managed Pod:

```bash
kubectl delete -f manifests/deployment.yaml
```

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- k3d
- kubectl
