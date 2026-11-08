# Exercise 2.5 — Documentation and ConfigMaps

## Overview

This exercise introduces Kubernetes ConfigMaps to manage application configuration separately from container images.

The Log-output application was updated to consume configuration from a ConfigMap.

The ConfigMap defines:

- A file named `information.txt`.
- An environment variable named `MESSAGE`.

The application reads both values and includes them in its HTTP response, alongside the existing timestamp, UUID and Ping-pong counter.

## ConfigMap

A ConfigMap named `log-output-config` was created in the `exercises` namespace.

Its data contains:

| Key               | Value                    |
| ----------------- | ------------------------ |
| `information.txt` | `this text is from file` |
| `MESSAGE`         | `hello world`            |

## Application changes

### Log-reader

The Log-reader was updated to:

- Read `information.txt` from `/usr/src/app/config/information.txt`.
- Read the `MESSAGE` environment variable.
- Include both configuration values in the HTTP response.
- Continue reading the timestamp and UUID from `log.txt`.
- Continue retrieving the Ping-pong counter through the internal Kubernetes Service.

A new Docker image was built and published:

`fjpiedade/log-reader:2.5`

### Log-writer

The Log-writer remains unchanged.

It continues generating a UUID and updating the timestamp every five seconds, writing the result to `log.txt`.

## Kubernetes configuration

The Log-output Deployment consumes the ConfigMap in two ways:

- **Volume mount:** the `information.txt` key is mounted as a file at `/usr/src/app/config/information.txt`.
- **Environment variable:** the `MESSAGE` key is injected into the Log-reader container through `configMapKeyRef`.

The ConfigMap volume is mounted as read-only.

The existing `emptyDir` volume remains shared between Log-writer and Log-reader for `log.txt`.

## Validation

The ConfigMap and Deployment manifests were validated using Kubernetes client-side dry runs.

The resources were applied to the `exercises` namespace.

The application was tested through the existing Ingress:

```bash
curl http://localhost:8081/
```

The response confirmed that both ConfigMap values were successfully consumed:

```text
file content: this text is from file
env variable: MESSAGE=hello world
2026-11-08T23:32:53.085963412Z: 1404908c-5089-42ba-8331-716d58c36bf6.
Ping / Pongs: 0
```

The timestamp, UUID and counter value vary between executions.

## Key concepts

- Kubernetes ConfigMaps.
- Configuration through environment variables.
- ConfigMaps mounted as volumes.
- `configMapKeyRef`.
- Volume mounts and read-only configuration.
- Separation of configuration from container images.
- Kubernetes Deployments and namespaces.
- Multi-container Pods.
