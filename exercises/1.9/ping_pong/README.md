# Exercise 1.9 – Ping-pong Application

Clone the repository:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes
```

## Description

This exercise introduces a second Go application called Ping-pong.

The application:

- Responds to `GET /pingpong` with `pong` and a counter.
- Increments the counter after every request.
- Stores the counter in memory.
- Runs on port `9901`.

## Docker

Build and publish the image:

```bash
cd exercises/1.9/ping_pong

docker build -t fjpiedade/ping-pong:1.0 .
docker push fjpiedade/ping-pong:1.0
```

## Kubernetes

The application runs in Kubernetes using a Deployment and a ClusterIP Service.

The shared Ingress routes requests as follows:

| Path        | Application |
| ----------- | ----------- |
| `/`         | Log output  |
| `/pingpong` | Ping-pong   |

Apply the manifests:

```bash
kubectl apply -f manifests/
```

**Note:** The Log output application and its Service must also be running in the cluster.

## Testing

Access the applications through the Ingress:

```bash
curl http://localhost:8081/
curl http://localhost:8081/pingpong
```

Expected Ping-pong responses:

```text
pong 0
pong 1
pong 2
```

The counter is stored in memory and may reset when the application restarts.
