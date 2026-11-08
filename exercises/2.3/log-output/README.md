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
- Retrieves the current Ping-pong counter through an HTTP GET request.
- Combines both values in its HTTP response.

The Log-reader communicates with Ping-pong using:

`http://ping-pong-svc:80/pings`

Kubernetes DNS resolves the Service name, so the Log-reader does not need to know the Ping-pong Pod IP.

## Shared storage

Both containers mount the same `emptyDir` volume at:

`/usr/src/app/files`

This allows the Log-writer and Log-reader to share `log.txt`.

The volume belongs to the Log-output Pod. It is not shared with the Ping-pong application.

## Expected response

```text
2026-11-07T07:52:52.743461856Z: 3e31de36-7a1f-45de-b8c7-59be37aac6ec.
Ping / Pongs: 1
```

The timestamp, UUID and counter value will vary.
