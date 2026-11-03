# Exercise 1.11 — Persisting data

## Architecture

This exercise uses one PersistentVolume (PV) and one PersistentVolumeClaim (PVC), shared by the `ping-pong` and `log-output` applications.

The shared volume contains two files:

- `pingpong.txt`: the persistent request counter, written by the Ping-pong application.
- `log.txt`: the UTC timestamp and UUID, written by `log-writer`.

The `log-writer` recovers the existing UUID from `log.txt` when it restarts. It updates the timestamp every five seconds while keeping the same UUID.

The `log-reader` reads both files and combines their contents in its HTTP response.

A request to `GET /` should return a response in this format:

```text
2020-03-30T12:15:17.705Z: 8523ecb1-c716-4cb6-a044-b9e83bb98e43.
Ping / Pongs: 3
```

## Prerequisites

- Git
- Go 1.25 or later
- Docker
- Kubernetes and kubectl
- A running k3d cluster

## Clone the repository

Clone the repository and move to its root directory:

```bash
git clone https://github.com/fjpiedade/DevOpsWithKubernetes.git
cd DevOpsWithKubernetes
```

All commands in this README are intended to be executed from the repository root unless stated otherwise.

## 1. Local testing

Before building Docker images, pushing them to Docker Hub, or deploying to Kubernetes, test all three Go applications locally.

This allows you to validate the shared-file behaviour and persistence without requiring containers or Kubernetes.

### Create the shared directory

The directory below simulates the shared PersistentVolume:

```bash
mkdir -p /tmp/dwk-1.11-data
```

Keep this directory between tests. It contains the files used by all three applications.

### Terminal 1 — Start Ping-pong

From the repository root:

```bash
cd exercises/1.11/ping-pong
COUNTER_FILE=/tmp/dwk-1.11-data/pingpong.txt PORT=9901 go run .
```

The Ping-pong application listens on port `9901`.

Open another terminal and test the endpoint:

```bash
curl http://localhost:9901/pingpong
```

Each request increments the counter. You should receive responses similar to:

```text
pong 1
pong 2
pong 3
```

### Terminal 2 — Start Log Writer

From the repository root, open a second terminal:

```bash
cd exercises/1.11/log-output/log-writer
LOG_FILE=/tmp/dwk-1.11-data/log.txt go run .
```

The Log Writer:

- Generates a UUID if `log.txt` does not exist.
- Saves the UUID and current UTC timestamp.
- Updates the timestamp every five seconds.
- Keeps the same UUID when restarted by recovering it from `log.txt`.

### Terminal 3 — Start Log Reader

From the repository root, open a third terminal:

```bash
cd exercises/1.11/log-output/log-reader
LOG_FILE=/tmp/dwk-1.11-data/log.txt COUNTER_FILE=/tmp/dwk-1.11-data/pingpong.txt PORT=9902 go run .
```

The Log Reader listens on port `9902`.

Test the combined response:

```bash
curl http://localhost:9902/
```

You should see the timestamp and UUID followed by the current Ping / Pongs counter.

You can also open the following address in your browser:

http://localhost:9902/

### Verify local persistence

Test that the applications recover their data after restarting.

**Ping-pong persistence**

1. Send several requests to `/pingpong`.
2. Note the current counter.
3. Stop the Ping-pong application using `Ctrl+C`.
4. Start it again using the same `COUNTER_FILE` path.
5. Send another request.

The counter should continue from its previous value instead of starting at zero.

**Log Writer persistence**

1. Note the UUID returned by the Log Reader.
2. Stop the Log Writer using `Ctrl+C`.
3. Start it again using the same `LOG_FILE` path.
4. Query the Log Reader again.

The UUID should remain unchanged. The timestamp should be updated.

Do not delete `/tmp/dwk-1.11-data` between these tests.

Once all local tests are successful, proceed with Docker and Kubernetes.

## 2. Build and push Docker images

Build the three images from the repository root:

```bash
docker build -t fjpiedade/ping-pong:1.1 exercises/1.11/ping-pong
docker build -t fjpiedade/log-writer:1.1 exercises/1.11/log-output/log-writer
docker build -t fjpiedade/log-reader:1.1 exercises/1.11/log-output/log-reader
```

Push the images to Docker Hub:

```bash
docker push fjpiedade/ping-pong:1.1
docker push fjpiedade/log-writer:1.1
docker push fjpiedade/log-reader:1.1
```

## 3. Prepare persistent storage in Kubernetes

The PV uses:

```yaml
hostPath: /tmp/kube
```

on the k3d node `k3d-k3s-default-agent-0`.

The directory must exist inside that node's container. Both Deployments use a `nodeSelector` to ensure that their Pods are scheduled on the same node.

First, verify the node name:

```bash
kubectl get nodes -o wide
```

Check that the directory exists inside the node container:

```bash
docker exec k3d-k3s-default-agent-0 ls -ld /tmp/kube
```

If the node name is different, replace `k3d-k3s-default-agent-0` in both files:

- `exercises/1.11/persistent-volume/persistent-volume.yaml`
- `exercises/1.11/manifests/deployment.yaml`

Make sure `/tmp/kube` exists inside the selected node.

### Apply the PV and PVC

Apply the persistent storage definitions first:

```bash
kubectl apply -f exercises/1.11/persistent-volume/persistent-volume.yaml
```

Check that the PV and PVC have been created:

```bash
kubectl get pv,pvc
```

The PVC `dwk-shared-pvc` should reach the `Bound` status before continuing.

## 4. Deploy the applications

Apply the Deployments:

```bash
kubectl apply -f exercises/1.11/manifests/deployment.yaml
```

Apply the Services:

```bash
kubectl apply -f exercises/1.11/manifests/service.yaml
```

Apply the Ingress:

```bash
kubectl apply -f exercises/1.11/manifests/ingress.yaml
```

If the old `log-output-ingress` from exercise 1.10 still exists, delete it to avoid keeping an obsolete Ingress rule:

```bash
kubectl delete ingress log-output-ingress --ignore-not-found
```

Check the Pods, Services, and Ingress:

```bash
kubectl get pods -o wide
kubectl get services
kubectl get ingress
```

## 5. Test the Kubernetes deployment

The k3d load balancer exposes the Ingress on port `8081`.

Test the Ping-pong endpoint:

```bash
curl http://localhost:8081/pingpong
```

Test the combined Log Output response:

```bash
curl http://localhost:8081/
```

Each request to `/pingpong` increments the persistent counter.

The `/` endpoint returns the latest timestamp and UUID, together with the current Ping / Pongs counter.

## 6. Verify persistence after Pod recreation

First, note the counter and UUID returned by:

```bash
curl http://localhost:8081/
```

Check the current Pods:

```bash
kubectl get pods -o wide
```

Delete both application Pods:

```bash
kubectl delete pod -l app=ping-pong
kubectl delete pod -l app=log-output
```

The Deployments should automatically create replacement Pods.

Wait until the new Pods are running:

```bash
kubectl get pods -o wide
```

Query the application again:

```bash
curl http://localhost:8081/
```

Verify that:

- The Ping / Pongs counter retains its previous value.
- The UUID remains unchanged.
- The timestamp may be newer because the Log Writer updates `log.txt` every five seconds.

This confirms that both applications are using the shared persistent volume.

## 7. Storage notes

The PV is static and uses `ReadWriteOnce`, `hostPath`, and node affinity.

Both Pods are pinned to the selected node so they can mount the same volume.

Keep both Deployments at one replica. The file-based counter is not designed for multiple concurrent Ping-pong replicas.

This setup is intended to preserve data when application Pods are recreated. It does not guarantee persistence if the k3d node or its container is deleted, because the data is stored inside that node's filesystem.
