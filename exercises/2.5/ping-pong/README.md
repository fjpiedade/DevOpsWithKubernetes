# Exercise 2.3 — Keep them separated

## Overview

This exercise introduces Kubernetes Namespaces to organize applications and keep their resources separated.

A namespace called `exercises` is created for the course applications. The Ping-pong and Log-output applications are deployed in this namespace instead of the default namespace.

The `default` namespace remains available for following the course material. Future exercises will use the `exercises` namespace, except for the project, which will have its own namespace.

## Architecture

The exercise contains two applications:

- **Ping-pong**: exposes the `/pingpong` and `/pings` HTTP endpoints.
- **Log-output**: runs a log-writer and a log-reader in the same Pod. The log-reader retrieves the Ping-pong counter through the Ping-pong Service.

Both applications run in the `exercises` namespace.

## Kubernetes Namespace

The namespace is defined in `namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: exercises
  labels:
    name: exercises
```

All application resources declare `namespace: exercises` in their metadata.

## Kubernetes resources

- `exercises` Namespace
- `ping-pong` Deployment
- `log-output` Deployment
- `ping-pong-svc` ClusterIP Service
- `log-output-svc` ClusterIP Service
- `persisting-data-ingress` Ingress

## Inter-pod communication

The Log-reader retrieves the Ping-pong counter using:

```text
http://ping-pong-svc:80/pings
```

Since both applications are in the same namespace, Kubernetes DNS resolves the Service name directly.

## Deployment

From the repository root:

Create the namespace:

```bash
kubectl apply -f exercises/2.3/namespace.yaml
```

Deploy the applications and networking resources:

```bash
kubectl apply -f exercises/2.3/manifests/
```

Check the resources:

```bash
kubectl get all -n exercises
kubectl get ingress -n exercises
```

## Testing

Check the Log-output application through the Ingress:

```bash
curl http://localhost:8081/
```

Check the Ping-pong endpoint:

```bash
curl http://localhost:8081/pingpong
```

Check the Ping-pong counter without incrementing it:

```bash
kubectl port-forward -n exercises svc/ping-pong-svc 8083:80
```

In another terminal:

```bash
curl http://localhost:8083/pings
```

The `/pings` endpoint must return the current counter, including immediately after the Ping-pong application starts.

## Container images

The applications use the following Docker Hub images:

- `fjpiedade/ping-pong:2.1`
- `fjpiedade/log-writer:2.1`
- `fjpiedade/log-reader:2.1`

The Ping-pong image retains the `2.1` tag. Its Docker Hub image was updated to initialize the counter file during application startup and allow `/pings` to return the initial counter value.

## Key concepts

- Kubernetes Namespaces
- Resource isolation and organization
- Deployments and Services
- ClusterIP
- Kubernetes internal DNS
- Ingress
- Communication between Pods in the same namespace

## Summary

The Ping-pong and Log-output applications are now deployed in the dedicated `exercises` namespace.

This establishes a consistent namespace strategy for future course exercises, while keeping the project separate.
