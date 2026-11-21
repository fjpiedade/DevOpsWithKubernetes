# Exercise 2.7 — Persisting Data with PostgreSQL

## Overview

This exercise extends the Ping-pong application by replacing file-based counter storage with a PostgreSQL database.

The application now stores and retrieves the Ping-pong counter directly from PostgreSQL, ensuring that the counter survives application and database Pod restarts.

The existing Log-output application is also retained, including its ConfigMap-based configuration.

## Architecture

The exercise contains three main components:

- **Ping-pong:** Go application exposing `/pingpong` and `/pings`.
- **PostgreSQL:** Database running as a StatefulSet with one replica.
- **Log-output:** Multi-container Pod running Log-writer and Log-reader.

All resources run in the `exercises` namespace.

## PostgreSQL

PostgreSQL runs using the `postgres:16-alpine` image.

Kubernetes resources:

- `postgres-secret` Secret containing database credentials.
- `postgres` headless Service.
- `postgres` StatefulSet with one replica.
- `postgres-data-postgres-0` PersistentVolumeClaim using the `local-path` StorageClass.

The database uses a persistent volume mounted at `/var/lib/postgresql/data`.

The application connects to PostgreSQL through the internal Kubernetes DNS name:

`postgres:5432`

## Ping-pong application

The Go application uses `github.com/lib/pq` to connect to PostgreSQL.

At startup, it creates the `pingpong_counter` table if it does not already exist and initializes the counter to zero when necessary.

The table contains:

| Column  | Description               |
| ------- | ------------------------- |
| `id`    | Primary key               |
| `count` | Current Ping-pong counter |

The application exposes two endpoints:

- `GET /pingpong`: atomically increments the counter and returns `pong N`.
- `GET /pings`: retrieves the current counter without incrementing it.

The counter is updated directly in PostgreSQL using `UPDATE ... RETURNING`.

## Log-output

The existing Log-output application remains deployed.

It continues to use the `log-output-config` ConfigMap for file-based and environment-variable configuration.

The Log-reader retrieves the current Ping-pong counter through the internal `ping-pong-svc` Kubernetes Service.

## Deployment

From the repository root:

Create the namespace:

```bash
kubectl apply -f exercises/2.7/namespace.yaml
```

Deploy PostgreSQL:

```bash
kubectl apply -f exercises/2.7/manifests/postgres.yaml
```

Deploy the application configuration and workloads:

```bash
kubectl apply -f exercises/2.7/manifests/configmap.yaml
kubectl apply -f exercises/2.7/manifests/deployment.yaml
kubectl apply -f exercises/2.7/manifests/service.yaml
kubectl apply -f exercises/2.7/manifests/ingress.yaml
```

Check the resources:

```bash
kubectl get pods,pvc -n exercises
kubectl get services,endpoints,ingress -n exercises
```

## Validation

The application was tested through the Ingress:

```bash
# Increment the Ping-pong counter
curl http://localhost:8081/pingpong

# Access the Log-output application
curl http://localhost:8081/
```

The Ping-pong counter can also be retrieved directly through its Kubernetes Service:

```bash
kubectl run ping-test --rm -i --restart=Never \
  --image=curlimages/curl:8.12.1 \
  -n exercises \
  -- curl -sS http://ping-pong-svc/pings
```

The PostgreSQL database was accessed directly:

```bash
kubectl exec -it postgres-0 -n exercises -- sh
psql -U pingpong -d pingpong
```

The counter was inspected using:

```sql
SELECT * FROM pingpong_counter;
```

A direct database update was performed:

```sql
UPDATE pingpong_counter
SET count = count + 1
WHERE id = 1
RETURNING count;
```

The updated value was immediately visible through the Log-output application, whose Log-reader retrieves the counter through the internal Ping-pong Service.

The counter was also verified to survive:

- Ping-pong Pod deletion and recreation.
- PostgreSQL Pod deletion and recreation.

This confirms that the counter is stored persistently in PostgreSQL using the StatefulSet's PersistentVolumeClaim.

## Container image

The Ping-pong application image:

`fjpiedade/ping-pong:2.7`

## Key concepts

- Kubernetes StatefulSets.
- PostgreSQL.
- PersistentVolumeClaims.
- Persistent storage.
- Kubernetes Secrets.
- Headless Services.
- Application-to-database connectivity.
- SQL queries and atomic updates.
- Data persistence across Pod restarts.
- Kubernetes Deployments, Services and Ingress.
