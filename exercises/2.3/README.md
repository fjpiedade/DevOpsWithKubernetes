# Exercise 2.1 — Connecting Pods

## Overview

This exercise connects the Ping-pong and Log-output applications using HTTP communication through Kubernetes Services and internal DNS.

Unlike Exercise 1.11, the two applications no longer share a PersistentVolumeClaim. The Log-reader retrieves the Ping-pong counter through an HTTP endpoint instead of reading a shared file.

## Architecture

### Ping-pong

Runs as an independent Kubernetes Deployment and exposes two HTTP GET endpoints on port `9901`:

- `GET /pingpong`: increments the counter and returns the current value as `pong N`.
- `GET /pings`: returns the current counter without incrementing it.

The counter is stored in `pingpong.txt` inside the Ping-pong container.

### Log-output

Runs as a Kubernetes Deployment with two containers in the same Pod:

- `log-writer`: generates a UUID and updates the UTC timestamp every five seconds.
- `log-reader`: exposes `GET /` on port `9901`.

Both containers share an `emptyDir` volume mounted at `/usr/src/app/files`, containing `log.txt`.

The Log-writer preserves the existing UUID when restarting by reading it from `log.txt`.

## Inter-pod communication

The Log-reader retrieves the current Ping-pong counter using HTTP GET:

`http://ping-pong-svc:80/pings`

Kubernetes internal DNS resolves `ping-pong-svc` to the Ping-pong Service.

The Service forwards requests from port `80` to the Ping-pong container on port `9901`.

The `/pings` endpoint only reads the counter; it does not increment it.

## Expected response

A request to the Log-output application:

`GET /`

Returns a response similar to:

```text
2026-11-07T07:52:52.743461856Z: 3e31de36-7a1f-45de-b8c7-59be37aac6ec.
Ping / Pongs: 1
```

The timestamp, UUID and counter value will vary.

## Storage

- Ping-pong stores its counter in `pingpong.txt` inside its container.
- Log-writer and Log-reader share `log.txt` through an `emptyDir` volume.
- There is no shared PV/PVC between Ping-pong and Log-output.
- The `emptyDir` volume is scoped to the Log-output Pod and is not shared with Ping-pong.
- The Ping-pong counter file is not backed by a PersistentVolume in this exercise, so it is not guaranteed to survive Pod recreation.

## Kubernetes resources

- Ping-pong Deployment
- Log-output Deployment
- `ping-pong-svc` ClusterIP Service
- `log-output-svc` ClusterIP Service
- Ingress for external HTTP access

## Key concepts

- Kubernetes Deployments
- Kubernetes Services
- ClusterIP
- Internal DNS
- HTTP communication between Pods
- Multi-container Pods
- `emptyDir` volumes
- Container-to-container communication
