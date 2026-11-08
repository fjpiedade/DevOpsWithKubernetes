# Log-output

The Log-output application consists of two containers running in the same Kubernetes Pod.

## Containers

### Log-writer

- Generates a UUID.
- Updates the UTC timestamp every five seconds.
- Writes the timestamp and UUID to `log.txt`.
- Recovers the existing UUID from `log.txt` when restarting.

### Log-reader

- Exposes `GET /` on port `9901`.
- Reads the timestamp and UUID from `log.txt`.
- Reads the contents of `information.txt`, mounted from a Kubernetes ConfigMap.
- Reads the `MESSAGE` environment variable, also provided by the ConfigMap.
- Retrieves the current Ping-pong counter through an HTTP GET request.
- Combines all values in its HTTP response.

## ConfigMap

The `log-output-config` ConfigMap is defined in the `exercises` namespace.

It provides two configuration entries:

| Key               | Value                    | Usage                               |
| ----------------- | ------------------------ | ----------------------------------- |
| `information.txt` | `this text is from file` | Mounted as a file                   |
| `MESSAGE`         | `hello world`            | Injected as an environment variable |

The `information.txt` file is mounted inside the Log-reader container at:

`/usr/src/app/config/information.txt`

The `MESSAGE` key is exposed through the `MESSAGE` environment variable.

This keeps application configuration separate from the container image.

## Ping-pong communication

The Log-reader communicates with Ping-pong using:

`http://ping-pong-svc:80/pings`

Kubernetes DNS resolves the Service name, so the Log-reader does not need to know the Ping-pong Pod IP.

The `/pings` endpoint retrieves the current counter without incrementing it.

## Shared storage

Both containers mount the same `emptyDir` volume at:

`/usr/src/app/files`

This allows the Log-writer and Log-reader to share `log.txt`.

The volume belongs to the Log-output Pod. It is not shared with the Ping-pong application.

## Expected response

A request to the Log-output application (`GET /`) returns a response similar to:

```text
file content: this text is from file
env variable: MESSAGE=hello world
2026-11-08T23:32:53.085963412Z: 1404908c-5089-42ba-8331-716d58c36bf6.
Ping / Pongs: 0
```

The timestamp, UUID and counter value will vary.

## Kubernetes resources

- Log-output Deployment with `log-writer` and `log-reader` containers.
- `log-output-svc` ClusterIP Service.
- `log-output-config` ConfigMap.
- `log-data` `emptyDir` volume shared between the containers.
- `config-volume` volume mounted from the ConfigMap.
- Ingress for external HTTP access.

## Key concepts

- Kubernetes ConfigMaps.
- Configuration through environment variables.
- ConfigMaps mounted as volumes.
- Multi-container Pods.
- Kubernetes Deployments and Services.
- Internal DNS.
- `emptyDir` volumes.
- Separation of application configuration from container images.
