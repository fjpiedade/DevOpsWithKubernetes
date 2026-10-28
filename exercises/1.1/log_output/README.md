# 1.1 - Log Output

A Go application that generates a random UUID at startup and prints it to standard output every five seconds, together with the current UTC timestamp.

This exercise covers application development with Go, Docker containerization, publishing an image to Docker Hub, and deploying the application to Kubernetes.

## Application behavior

- Generates a UUID when the application starts.
- Keeps the same UUID for the lifetime of the process.
- Prints the current UTC timestamp and UUID every five seconds.
- Writes logs to standard output so they can be inspected using Docker or Kubernetes logging commands.

Example output:

```text
2026-10-28T00:52:46.877Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
2026-10-28T00:52:51.880Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
2026-10-28T00:52:56.886Z: 068901c3-3554-4d9d-a28f-0d30d8501ff9
```

The UUID remains the same while the timestamps change.

## Run on Kubernetes

### Prerequisites

- A running Kubernetes cluster, such as k3d.
- `kubectl` configured to access the cluster.

The application image is publicly available on [Docker Hub](https://hub.docker.com/r/fjpiedade/log-output). You can deploy it directly without cloning this repository or building the image.

Create a Deployment using the published image:

```bash
kubectl create deployment log-output --image=fjpiedade/log-output:1.0
```

Check the Deployment status:

```bash
kubectl get deployments
kubectl get pods
```

Wait for the Deployment to become available:

```bash
kubectl rollout status deployment/log-output
```

View the application logs:

```bash
kubectl logs -f deployment/log-output
```

The same UUID should appear every five seconds, with a new UTC timestamp on each line.

Press `Ctrl+C` to stop following the logs. The application continues running in Kubernetes.

To remove the Deployment when finished:

```bash
kubectl delete deployment log-output
```

## Run with Docker

Run the published image directly:

```bash
docker run --rm fjpiedade/log-output:1.0
```

Press `Ctrl+C` to stop the container.

## Run locally with Go

To run the application directly on your machine, Go must be installed.

Clone the repository and navigate to the application directory:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes/exercises/1.1/log_output
```

## Kubernetes

The application is deployed to Kubernetes using a declarative Deployment manifest located at `manifests/deployment.yaml`.

Apply the manifest:

```bash
kubectl apply -f exercises/1.1/log_output/manifests/deployment.yaml
```

Check the Deployment and Pods:

```bash
kubectl rollout status deployment/log-output
kubectl get deployments
kubectl get pods
```

Follow the application logs:

```bash
kubectl logs -f deployment/log-output
```

Restart the Deployment and verify that the application starts with a new UUID:

```bash
kubectl rollout restart deployment/log-output
kubectl rollout status deployment/log-output
kubectl logs -f deployment/log-output
```

The application logs a UTC timestamp and its UUID every five seconds. The UUID remains the same while the application instance is running and changes after a restart.

Run the application:

```bash
go run .
```

Press `Ctrl+C` to stop the application.

## Project structure

```text
log_output/
├── .dockerignore
├── Dockerfile
├── go.mod
├── main.go
└── README.md
```

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- k3d
- kubectl
