# Exercise 1.10 – Even More Services

## Overview

The goal of this exercise is to split the Log Output application into two separate containers running inside the same Kubernetes Pod.

The application consists of:

- **Log Writer:** Generates a random UUID when it starts and writes the UUID together with a UTC timestamp to a shared file every 5 seconds.
- **Log Reader:** Reads the shared file and exposes its contents through an HTTP GET endpoint.

Both containers share an `emptyDir` volume, allowing them to access the same file.

## Architecture

The application runs inside a single Kubernetes Pod with two containers:

```text
                    Kubernetes Pod
                 ┌───────────────────────┐
                 │                       │
                 │    log-writer         │
                 │    Writes every 5s   │
                 │          │            │
                 │          ▼            │
                 │   ┌───────────────┐   │
                 │   │   emptyDir    │   │
                 │   │   log.txt     │   │
                 │   └───────────────┘   │
                 │          │            │
                 │          ▼            │
                 │    log-reader         │
                 │    HTTP :9901        │
                 │                       │
                 └───────────┬───────────┘
                             │
                       Service :80
                             │
                          Ingress
                             │
                   http://localhost:8081/
```

## Project Structure

```text
log-output/
├── log-writer/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── log-reader/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
└── manifests/
    ├── deployment.yaml
    ├── service.yaml
    └── ingress.yaml
```

## Docker Images

The two applications are built and published as separate Docker images:

| Application | Docker Image               |
| ----------- | -------------------------- |
| Log Writer  | `fjpiedade/log-writer:1.0` |
| Log Reader  | `fjpiedade/log-reader:1.0` |

### Build and Push

Build the Log Writer:

```bash
cd log-writer
docker build -t fjpiedade/log-writer:1.0 .
docker push fjpiedade/log-writer:1.0
```

Build the Log Reader:

```bash
cd ../log-reader
docker build -t fjpiedade/log-reader:1.0 .
docker push fjpiedade/log-reader:1.0
```

## Kubernetes Deployment

The Deployment creates one Pod containing both containers.

An `emptyDir` volume named `shared-log` is mounted by both containers at:

```text
/usr/src/app/files
```

Both applications use the same file:

```text
/usr/src/app/files/log.txt
```

Apply the Kubernetes manifests:

```bash
kubectl apply -f manifests/
```

Check the Deployment and Pods:

```bash
kubectl get deployments
kubectl get pods -l app=log-output
```

The Pod should show `2/2 Running`.

### Inspect Container Logs

Check the Log Writer:

```bash
kubectl logs deployment/log-output -c log-writer
```

Check the Log Reader:

```bash
kubectl logs deployment/log-output -c log-reader
```

## Accessing the Application

The `log-output-svc` Service exposes the Log Reader internally on port `80`, forwarding requests to container port `9901`.

The `log-output-ingress` Ingress exposes the application externally.

Access the application using:

```bash
curl http://localhost:8081/
```

The response contains timestamped entries similar to:

```text
2026-11-01T22:23:04.526Z: 57bc881e-fef4-40d8-87e5-261dfdbcac81
2026-11-01T22:23:09.530Z: 57bc881e-fef4-40d8-87e5-261dfdbcac81
2026-11-01T22:23:14.534Z: 57bc881e-fef4-40d8-87e5-261dfdbcac81
```

The UUID remains the same because it is generated once when the Log Writer starts.

## Cleanup

Remove the Kubernetes resources:

```bash
kubectl delete -f manifests/
```

## Key Concepts

This exercise demonstrates:

- Running multiple containers inside a single Kubernetes Pod.
- Sharing files between containers using an `emptyDir` volume.
- Separating application responsibilities into independent containers.
- Exposing an application through a Kubernetes Service.
- Routing external HTTP traffic using an Ingress.
- Inspecting logs from individual containers using `kubectl logs -c`.
