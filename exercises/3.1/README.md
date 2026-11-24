# Exercise 3.1 — Pingpong GKE

## Overview

This exercise deploys the Ping-pong application to Google Kubernetes Engine (GKE).

The application is exposed to the internet using a Kubernetes `LoadBalancer` Service.

PostgreSQL is deployed as a StatefulSet with persistent storage to retain the Ping-pong counter across application and database Pod restarts.

## Architecture

The exercise consists of three main components:

- **Ping-pong:** Go application exposing `/pingpong` and `/pings`.
- **PostgreSQL:** Database running as a StatefulSet with one replica and persistent storage.
- **LoadBalancer:** Kubernetes Service exposing the Ping-pong application externally.

All Kubernetes resources run in the `exercises` namespace.

The request flow is:

```text
Internet
   |
   v
LoadBalancer Service (port 80)
   |
   v
Ping-pong Deployment (port 9901)
   |
   v
PostgreSQL Service (port 5432)
   |
   v
PostgreSQL StatefulSet
```

## GKE cluster

The application was deployed to a GKE cluster with the following configuration:

| Setting            | Value                |
| ------------------ | -------------------- |
| Cluster            | `dwk-cluster`        |
| Project            | `dwk-gke-fjp-2026`   |
| Zone               | `europe-north1-b`    |
| Kubernetes version | `1.36.6-gke.1146000` |
| Machine type       | `e2-small`           |
| Number of nodes    | 4                    |

## PostgreSQL

PostgreSQL runs using the `postgres:16-alpine` image.

Kubernetes resources:

- `postgres-secret`: Secret containing database credentials.
- `postgres`: Headless Service.
- `postgres`: StatefulSet with one replica.
- `postgres-data-postgres-0`: PersistentVolumeClaim with 1 GiB of storage.

The PersistentVolumeClaim uses the GKE `standard-rwo` StorageClass, backed by Google Persistent Disk through the GKE CSI driver.

The database volume is mounted at:

`/var/lib/postgresql/data`

The Ping-pong application connects to PostgreSQL using the internal Kubernetes DNS name:

`postgres:5432`

## Ping-pong application

The Go application uses `github.com/lib/pq` to connect to PostgreSQL.

At startup, it creates the `pingpong_counter` table if it does not already exist and initializes the counter to zero when necessary.

The application exposes two endpoints:

- `GET /pingpong`: Atomically increments the counter and returns `pong N`.
- `GET /pings`: Retrieves the current counter without incrementing it.

The counter is stored in PostgreSQL using an `UPDATE ... RETURNING` query.

The application runs in a Kubernetes Deployment with one replica.

Container image:

`fjpiedade/ping-pong:2.7`

Container port: `9901`.

Database connection settings are provided through environment variables. Database credentials are retrieved from the `postgres-secret` Kubernetes Secret.

## LoadBalancer Service

The Ping-pong application is exposed using a Kubernetes Service of type `LoadBalancer`.

Service configuration:

| Setting       | Value           |
| ------------- | --------------- |
| Name          | `ping-pong-svc` |
| Type          | `LoadBalancer`  |
| Namespace     | `exercises`     |
| External port | `80`            |
| Target port   | `9901`          |

GKE automatically provisions an external IP address for the LoadBalancer.

The external IP assigned during validation was:

`35.228.115.104`

The IP address may change if the Service or cluster is recreated.

## Deployment

From the repository root:

Create the namespace:

```bash
kubectl apply -f exercises/3.1/namespace.yaml
```

Deploy PostgreSQL:

```bash
kubectl apply -f exercises/3.1/manifests/postgres.yaml
```

Deploy the Ping-pong application:

```bash
kubectl apply -f exercises/3.1/manifests/ping-pong-deployment.yaml
```

Expose the application using a LoadBalancer:

```bash
kubectl apply -f exercises/3.1/manifests/ping-pong-loadbalancer.yaml
```

Check the Pods and persistent volume:

```bash
kubectl get pods,pvc -n exercises
```

Check the LoadBalancer and its external IP:

```bash
kubectl get svc ping-pong-svc -n exercises
```

Wait for the external IP to be assigned if its status is initially `pending`.

## Validation

The application was tested through its external LoadBalancer IP.

Increment the counter:

```bash
curl http://<EXTERNAL-IP>/pingpong
```

Retrieve the current counter:

```bash
curl http://<EXTERNAL-IP>/pings
```

During validation, the application returned:

```text
pong 3
```

And:

```text
3
```

The counter had also been incremented through browser requests.

The following Kubernetes resources were verified:

- Ping-pong Pod: `1/1 Running`.
- PostgreSQL Pod: `1/1 Running`.
- PostgreSQL PersistentVolumeClaim: `Bound`.
- LoadBalancer Service: `EXTERNAL-IP` assigned and reachable.

## Key concepts

- Google Kubernetes Engine (GKE).
- Kubernetes Deployments.
- Kubernetes StatefulSets.
- Kubernetes Services of type `LoadBalancer`.
- External access to Kubernetes applications.
- PostgreSQL.
- PersistentVolumeClaims.
- GKE StorageClasses.
- Google Persistent Disk CSI driver.
- Kubernetes Secrets.
- Application-to-database connectivity.
- SQL queries and atomic counter updates.
