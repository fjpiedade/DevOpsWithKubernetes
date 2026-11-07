# Ping-pong

HTTP service that maintains a counter of Ping-pong requests in `pingpong.txt`.

The application exposes two GET endpoints on port `9901`:

- `GET /pingpong`: increments the counter and returns the current value as `pong N`.
- `GET /pings`: returns the current counter without incrementing it.

The `/pings` endpoint is used by the Log-reader to retrieve the current counter through HTTP.

## Kubernetes

The application runs in its own Deployment and is exposed internally through the `ping-pong-svc` ClusterIP Service.

Other Pods can reach it using:

`http://ping-pong-svc:80`

The Service forwards requests to container port `9901`.

The counter file is stored inside the container filesystem. It is not backed by a PersistentVolume in Exercise 2.1.
