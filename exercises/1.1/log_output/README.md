# 1.1 - Log Output

A Go application that generates a random UUID at startup and prints it to standard output every five seconds, together with the current UTC timestamp.

This exercise covers application development with Go, Docker containerization, publishing images to Docker Hub, and deploying an application to Kubernetes using k3d.

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

The application image is publicly available on [Docker Hub](https://hub.docker.com/r/fjpiedade/log-output). You can run it directly without cloning this repository or building the image.

### Option 1: Run directly using the published image

Create a standalone Pod:

```bash
kubectl run log-output --image=fjpiedade/log-output:1.0 --restart=Never
```

Check the Pod status:

```bash
kubectl get pods
```

View the application logs:

```bash
kubectl logs -f log-output
```

The same UUID should appear every five seconds, with a new UTC timestamp on each line.

Press `Ctrl+C` to stop following the logs. The Pod continues running.

Delete the Pod when finished:

```bash
kubectl delete pod log-output
```

### Option 2: Deploy using a Kubernetes manifest

If you want to use the Deployment configuration included in this repository, clone the repository first.

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes/exercises/1.1/log_output
```

Apply the Deployment manifest:

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

Press `Ctrl+C` to stop following the logs. The application continues running in Kubernetes.

To remove the Deployment and its managed Pod:

```bash
kubectl delete -f manifests/deployment.yaml
```

A Deployment is generally preferable when you want Kubernetes to manage the desired number of replicas and recreate Pods when necessary.

## Run with Docker

### Run the published image

Docker downloads the image from Docker Hub if it is not already available locally.

```bash
docker run --rm fjpiedade/log-output:1.0
```

Alternatively, run the `latest` tag:

```bash
docker run --rm fjpiedade/log-output:latest
```

If you omit the tag, Docker defaults to `latest`:

```bash
docker run --rm fjpiedade/log-output
```

Press `Ctrl+C` to stop the container.

### Build the image locally

To build the image yourself, clone the repository and navigate to the application directory:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes/exercises/1.1/log_output
```

Build the image:

```bash
docker build -t log-output:1.0 .
```

Run the container:

```bash
docker run --rm --name log-output-test log-output:1.0
```

Press `Ctrl+C` to stop the container.

## Run locally with Go

To run the application directly on your machine, you need to have Go installed.

Clone the repository if you have not already done so, navigate to `exercises/1.1/log_output`, and run:

```bash
go run .
```

Press `Ctrl+C` to stop the application.

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

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- k3d
- kubectl
