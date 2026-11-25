# Exercise 3.3 — To the Gateway

Deploy the **Log output** and **Ping-pong** applications to Google Kubernetes Engine (GKE) and expose them through the Kubernetes Gateway API, replacing the Ingress used in Exercise 3.2.

## Architecture

The exercise runs in the `exercises` namespace.

- **Log output** runs as a Deployment with two containers:
  - `log-writer` periodically writes a timestamp and UUID to a shared file.
  - `log-reader` serves the file contents, ConfigMap information, the `MESSAGE` environment variable, and the Ping-pong counter.
- **Ping-pong** runs as a Deployment and stores its counter in PostgreSQL.
- **PostgreSQL** runs as a StatefulSet with a persistent volume.
- **Gateway API** exposes the applications through a GKE-managed external HTTP Load Balancer.
- **HTTPRoute** defines the path-based routing rules.

Both application Services use `ClusterIP`, since the Gateway routes traffic to Services inside the cluster.

### Gateway routes

| Path        | Service          | Purpose               |
| ----------- | ---------------- | --------------------- |
| `/`         | `log-output-svc` | Log output            |
| `/pingpong` | `ping-pong-svc`  | Increment the counter |
| `/pings`    | `ping-pong-svc`  | Read the counter      |

The Ping-pong application also responds with `200 OK` at `/`. This endpoint is used by its Kubernetes readiness and liveness probes. Public requests to `/` are routed to Log output.

## Container images

This exercise reuses the images published for Exercise 3.2, because the application source code has not changed:

- `fjpiedade/ping-pong:3.2`
- `fjpiedade/log-reader:3.2`
- `fjpiedade/log-writer:3.2`

No new image build or push is required.

## Deploy to GKE

### 1. Create a GKE cluster with Gateway API enabled

Create the cluster with the Gateway API Standard channel enabled:

```bash
gcloud container clusters create dwk-cluster \
  --project dwk-gke-fjp-2026 \
  --zone europe-north1-b \
  --disk-size 32 \
  --num-nodes 1 \
  --machine-type e2-small \
  --gateway-api=standard
```

Configure `kubectl` credentials:

```bash
gcloud container clusters get-credentials dwk-cluster \
  --project dwk-gke-fjp-2026 \
  --zone europe-north1-b
```

Verify that the GKE GatewayClass is available:

```bash
kubectl get gatewayclass
```

The `gke-l7-global-external-managed` GatewayClass should be accepted.

### 2. Create the namespace and deploy the applications

Create the namespace:

```bash
kubectl apply -f exercises/3.3/namespace.yaml
```

Deploy the ConfigMap and PostgreSQL:

```bash
kubectl apply -f exercises/3.3/manifests/configmap.yaml
kubectl apply -f exercises/3.3/manifests/postgres.yaml

kubectl rollout status statefulset/postgres -n exercises
```

Deploy Ping-pong and Log output:

```bash
kubectl apply -f exercises/3.3/manifests/ping-pong-deployment.yaml
kubectl apply -f exercises/3.3/manifests/log-output-deployment.yaml

kubectl rollout status deployment/ping-pong -n exercises
kubectl rollout status deployment/log-output -n exercises
```

Create the application Services:

```bash
kubectl apply -f exercises/3.3/manifests/services.yaml
```

### 3. Create the Gateway and HTTPRoute

Create the Gateway:

```bash
kubectl apply -f exercises/3.3/manifests/gateway.yaml
```

Create the HTTPRoute:

```bash
kubectl apply -f exercises/3.3/manifests/route.yaml
```

The Gateway uses the `gke-l7-global-external-managed` GatewayClass. The HTTPRoute attaches to this Gateway and directs requests to the appropriate application Service.

Check the Gateway and route status:

```bash
kubectl get gateway -n exercises
kubectl get httproute -n exercises
```

Wait until the Gateway reports `PROGRAMMED=True` and the HTTPRoute reports `Accepted=True` and `ResolvedRefs=True`. Provisioning the Google Cloud Load Balancer can take several minutes.

## Verify the deployment

Check the workloads and Services:

```bash
kubectl get all -n exercises
kubectl get pods -n exercises -o wide
```

Get the external IP address assigned to the Gateway:

```bash
GATEWAY_IP=$(kubectl get gateway my-gateway \
  -n exercises \
  -o jsonpath='{.status.addresses[0].value}')

echo "$GATEWAY_IP"
```

Test the public endpoints:

```bash
curl -i "http://$GATEWAY_IP/"
curl -i "http://$GATEWAY_IP/pingpong"
curl -i "http://$GATEWAY_IP/pings"
```

Expected behaviour:

- `/` returns the Log output information.
- `/pingpong` increments the counter and returns the new value.
- `/pings` returns the current counter without incrementing it.

## Cleanup

The GKE Gateway provisions Google Cloud Load Balancer resources, which can incur costs. Delete the HTTPRoute and Gateway first, and allow the associated cloud resources to be removed.

```bash
kubectl delete -f exercises/3.3/manifests/route.yaml
kubectl delete -f exercises/3.3/manifests/gateway.yaml
```

Check that the Gateway has been deleted:

```bash
kubectl get gateway -n exercises
```

Then remove the exercise namespace and its remaining resources:

```bash
kubectl delete namespace exercises
```

If the GKE cluster was created specifically for this exercise and is no longer needed, delete it:

```bash
gcloud container clusters delete dwk-cluster \
  --project dwk-gke-fjp-2026 \
  --zone europe-north1-b
```

Finally, check Google Cloud for any remaining Load Balancer resources or persistent disks to avoid ongoing costs.
