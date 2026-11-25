# Exercise 3.2 — Back to Ingress

Deploy the **Log output** and **Ping-pong** applications to Google Kubernetes Engine (GKE) and expose them through a Kubernetes Ingress.

## Architecture

The exercise runs in the `exercises` namespace.

- **Log output** runs as a Deployment with two containers:
  - `log-writer` periodically writes a timestamp and UUID to a shared file.
  - `log-reader` serves the file contents, the ConfigMap information, the `MESSAGE` environment variable, and the Ping-pong counter.
- **Ping-pong** runs as a Deployment and stores its counter in PostgreSQL.
- **PostgreSQL** runs as a StatefulSet with a persistent volume.
- **Ingress** routes external HTTP requests to the application Services.

### Ingress routes

| Path        | Service          | Purpose               |
| ----------- | ---------------- | --------------------- |
| `/`         | `log-output-svc` | Log output            |
| `/pingpong` | `ping-pong-svc`  | Increment the counter |
| `/pings`    | `ping-pong-svc`  | Read the counter      |

The Ping-pong application also responds with `200 OK` at `/`. This root endpoint is used by Kubernetes readiness/liveness probes and by the GKE Ingress health check.

## Container images

The following images are published on Docker Hub:

- `fjpiedade/ping-pong:3.2`
- `fjpiedade/log-reader:3.2`
- `fjpiedade/log-writer:3.2`

To rebuild and push them from the repository root:

```bash
docker build -t fjpiedade/ping-pong:3.2 exercises/3.2/ping-pong
docker build -t fjpiedade/log-reader:3.2 exercises/3.2/log-output/log-reader
docker build -t fjpiedade/log-writer:3.2 exercises/3.2/log-output/log-writer

docker push fjpiedade/ping-pong:3.2
docker push fjpiedade/log-reader:3.2
docker push fjpiedade/log-writer:3.2
```

## Deploy to GKE

Make sure `kubectl` is connected to the intended GKE cluster.

Create the namespace:

```bash
kubectl apply -f exercises/3.2/namespace.yaml
```

Deploy the ConfigMap and PostgreSQL:

```bash
kubectl apply -f exercises/3.2/manifests/configmap.yaml
kubectl apply -f exercises/3.2/manifests/postgres.yaml

kubectl rollout status statefulset/postgres -n exercises
```

Deploy Ping-pong and the Services:

```bash
kubectl apply -f exercises/3.2/manifests/ping-pong-deployment.yaml
kubectl apply -f exercises/3.2/manifests/services.yaml

kubectl rollout status deployment/ping-pong -n exercises
```

Deploy Log output:

```bash
kubectl apply -f exercises/3.2/manifests/log-output-deployment.yaml

kubectl rollout status deployment/log-output -n exercises
```

Finally, create the Ingress:

```bash
kubectl apply -f exercises/3.2/manifests/ingress.yaml
kubectl get ingress -n exercises
```

The GKE Ingress controller uses the `kubernetes.io/ingress.class: "gce"` annotation in the Ingress manifest.

Wait for the Ingress to receive an external IP address. Provisioning the Google Cloud Load Balancer can take several minutes.

## Verify the deployment

Check the workloads and Services:

```bash
kubectl get all -n exercises
kubectl get ingress -n exercises
```

Get the external IP from the Ingress:

```bash
INGRESS_IP=$(kubectl get ingress log-output-ping-pong-ingress \
  -n exercises \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "$INGRESS_IP"
```

Test the public endpoints:

```bash
curl "http://$INGRESS_IP/"
curl "http://$INGRESS_IP/pingpong"
curl "http://$INGRESS_IP/pings"
```

Expected behaviour:

- `/` returns the Log output information.
- `/pingpong` increments the counter and returns the new value.
- `/pings` returns the current counter without incrementing it.

## Cleanup

The Ingress provisions Google Cloud Load Balancer resources, which can incur costs. Delete the Ingress first and allow its associated cloud resources to be removed.

```bash
kubectl delete -f exercises/3.2/manifests/ingress.yaml
```

Then remove the exercise namespace and its resources:

```bash
kubectl delete namespace exercises
```

If the GKE cluster was created specifically for this exercise and is no longer needed, delete it as well:

```bash
gcloud container clusters delete dwk-cluster \
  --project dwk-gke-fjp-2026 \
  --zone europe-north1-b
```

Check Google Cloud for any remaining load balancer resources or disks after cleanup.
