# Log Output

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
kubectl create deployment log-output --image=fjpiedade/log-output:1.1
```

## Run locally with Go

### Create the k3d cluster

```bash
k3d cluster create --port 8082:30080@agent:0 -p 8081:80@loadbalancer --agents 2
```

To run the application directly on your machine, Go must be installed.

Clone the repository and navigate to the application directory:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes/exercises/1.1/log_output
```

# Exercise 1.7 - External access with Ingress

The Log Output application is exposed outside the Kubernetes cluster using an Ingress resource.

The application generates a UUID at startup and prints the current UTC timestamp and UUID to standard output every five seconds.

An HTTP endpoint is also available to retrieve the current timestamp and UUID.

## Application behavior

- Generates a UUID when the application starts.
- Keeps the same UUID for the lifetime of the process.
- Prints the timestamp and UUID to the logs every five seconds.
- Exposes an HTTP GET endpoint at `/`.
- Returns the current UTC timestamp and UUID through HTTP.

## Kubernetes resources

The application uses the following Kubernetes resources:

- **Deployment:** Runs the Log Output application.
- **Service:** Exposes the application internally using ClusterIP.
- **Ingress:** Provides external HTTP access through Traefik.

## Apply the Kubernetes manifests

From the repository root:

```bash
kubectl apply -f exercises/1.1/log_output/manifests/
```

## Verify the resources

Check the Deployment:

```bash
kubectl get deployments
kubectl rollout status deployment/log-output
```

Check the Pods:

```bash
kubectl get pods
```

Check the Service:

```bash
kubectl get services
kubectl describe service log-output
```

Check the Ingress:

```bash
kubectl get ingress
kubectl describe ingress log-output
```

## Access the application

The k3d cluster exposes the Traefik Ingress Controller through port `8081` on the host.

Open the application in a browser:

```text
http://localhost:8081/
```

Alternatively, test the endpoint using curl:

```bash
curl http://localhost:8081/
```

Expected response:

```text
2026-10-30T22:12:18.887Z: 395bedf4-6814-42e7-8ca7-926e8fe1bfc1
```

The timestamp changes with each request, while the UUID remains the same during the application's lifetime.

## Application logs

Follow the application logs:

```bash
kubectl logs -f deployment/log-output
```

The application prints a new UTC timestamp every five seconds, keeping the same UUID.

## Architecture

External HTTP traffic follows this path:

```text
Browser
   |
   v
k3d Load Balancer (port 8081)
   |
   v
Traefik Ingress Controller
   |
   v
Ingress (log-output)
   |
   v
Service (log-output:80)
   |
   v
Pod (log-output:9901)
```

## Technologies

- Go
- Docker
- Docker Hub
- Kubernetes
- k3d
- Traefik
- kubectl
