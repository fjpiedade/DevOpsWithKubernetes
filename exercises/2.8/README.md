# Todo App — Exercise 2.8

## Overview

This exercise continues the Todo application developed in previous exercises.

The main objective is to persist todos in a PostgreSQL database running inside Kubernetes.

The database is deployed as a StatefulSet with one replica and uses persistent storage through a PersistentVolumeClaim. The backend connects to PostgreSQL using configuration provided through Kubernetes ConfigMaps and Secrets.

The application runs in the dedicated `project` namespace, keeping project-related resources separate from regular course exercises in the `exercises` namespace.

## Architecture

The project consists of two Go services and a PostgreSQL database:

- **todo-app**: serves the web interface, static assets and cached random image. It also acts as a reverse proxy for Todo API requests.
- **todo-backend**: provides a REST API to retrieve and create todos. Todos are stored in PostgreSQL.
- **todo-postgres**: PostgreSQL database deployed as a StatefulSet with one replica.

The frontend communicates with the backend through `/todos`. The backend reads and writes todos in PostgreSQL.

The application also uses persistent storage for the cached image.

```text
Browser
   |
   v
Ingress (Traefik)
   |
   v
todo-app-svc
   |
   v
todo-app (Go)
   |
   | Reverse proxy: /todos
   v
todo-backend-svc
   |
   v
todo-backend (Go)
   |
   v
todo-postgres (Headless Service)
   |
   v
PostgreSQL StatefulSet (1 replica)
   |
   v
PersistentVolumeClaim
```

Both Ingress paths, `/` and `/todos`, are routed to `todo-app-svc`. The frontend forwards API requests to the backend using `TODO_BACKEND_URL`.

The backend connects to PostgreSQL through the Kubernetes DNS name `todo-postgres`.

## Kubernetes Namespace

All namespaced application resources are deployed in the `project` namespace.

The namespace is defined in `namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: project
```

The `project` namespace is separate from the `exercises` namespace.

PersistentVolumes are cluster-scoped resources and do not belong to a namespace. PersistentVolumeClaims belong to `project`.

## Kubernetes Resources

- `project` Namespace
- `todo-app` Deployment
- `todo-backend` Deployment
- `todo-postgres` StatefulSet (1 replica)
- `todo-app-svc` ClusterIP Service
- `todo-backend-svc` ClusterIP Service
- `todo-postgres` Headless Service
- `todo-app-ingress` Ingress
- `todo-app-config` ConfigMap
- `todo-backend-config` ConfigMap
- `todo-postgres-secret` Secret
- `persisting-data-pv` PersistentVolume
- `persisting-data-pvc` PersistentVolumeClaim
- PostgreSQL PersistentVolumeClaim created through the StatefulSet `volumeClaimTemplates`

## Configuration Management

Runtime configuration is provided through Kubernetes ConfigMaps and Secrets.

### todo-app-config

The `todo-app-config` ConfigMap contains:

| Variable               | Description                    | Value                        |
| ---------------------- | ------------------------------ | ---------------------------- |
| `PORT`                 | Application listening port     | `9901`                       |
| `IMAGE_FILE`           | Cached image file location     | `/data/image.jpg`            |
| `TODO_BACKEND_URL`     | Backend Kubernetes Service URL | `http://todo-backend-svc:80` |
| `IMAGE_SOURCE_URL`     | External image source          | `https://picsum.photos/1200` |
| `IMAGE_CACHE_DURATION` | Image cache duration           | `10m`                        |
| `IMAGE_HTTP_TIMEOUT`   | HTTP client timeout            | `30s`                        |

### todo-backend-config

The `todo-backend-config` ConfigMap contains:

| Variable  | Description                 | Value           |
| --------- | --------------------------- | --------------- |
| `PORT`    | Backend listening port      | `8080`          |
| `DB_HOST` | PostgreSQL Service hostname | `todo-postgres` |
| `DB_PORT` | PostgreSQL listening port   | `5432`          |
| `DB_NAME` | Database name               | `todos`         |

### todo-postgres-secret

The `todo-postgres-secret` Secret contains the PostgreSQL credentials and database configuration:

| Key                 | Description                                      |
| ------------------- | ------------------------------------------------ |
| `POSTGRES_USER`     | PostgreSQL username                              |
| `POSTGRES_PASSWORD` | PostgreSQL password                              |
| `POSTGRES_DB`       | Database initialized by the PostgreSQL container |

The PostgreSQL StatefulSet reads these values from the Secret.

The backend maps `POSTGRES_USER` and `POSTGRES_PASSWORD` to its `DB_USER` and `DB_PASSWORD` environment variables.

The credentials in the manifest are intended for this development exercise only.

## PostgreSQL and Persistent Storage

PostgreSQL runs as a StatefulSet named `todo-postgres` with exactly one replica.

The StatefulSet uses a headless Service named `todo-postgres` for stable network identity.

A `volumeClaimTemplates` entry creates a PersistentVolumeClaim for the PostgreSQL data directory:

- Claim name: `postgres-data-todo-postgres-0`
- Access mode: `ReadWriteOnce`
- Storage class: `local-path`
- Requested storage: `1Gi`

The PostgreSQL container stores its database files under `/var/lib/postgresql/data`.

The backend automatically creates the `todos` table if it does not already exist:

```sql
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    text VARCHAR(140) NOT NULL
);
```

The table stores each todo with an automatically generated ID and its text.

The PostgreSQL PersistentVolumeClaim keeps database data across PostgreSQL Pod restarts.

The frontend also uses a separate PersistentVolume and PersistentVolumeClaim for its cached image:

- `persisting-data-pv`
- `persisting-data-pvc`

These two storage configurations serve different purposes: one persists PostgreSQL data, while the other stores the frontend image cache.

## Deployment

Run the following commands from the `exercises/2.8/` directory.

### 1. Create the namespace

```bash
kubectl apply -f namespace.yaml
```

### 2. Apply persistent storage for the frontend

```bash
kubectl apply -f persistent-volume/
```

Check the storage resources:

```bash
kubectl get pv
kubectl get pvc -n project
```

The frontend PersistentVolumeClaim should reach the `Bound` status.

The PostgreSQL PersistentVolumeClaim is created automatically when the StatefulSet is deployed.

### 3. Apply ConfigMaps

```bash
kubectl apply -f manifests/configmap.yaml
```

### 4. Deploy PostgreSQL

```bash
kubectl apply -f manifests/postgres.yaml
```

Wait for the database Pod to become ready:

```bash
kubectl rollout status statefulset/todo-postgres -n project --timeout=180s
```

Check the database:

```bash
kubectl exec -n project todo-postgres-0 -- sh -c \
  'pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
```

### 5. Build and publish the backend image

```bash
docker build -t fjpiedade/todo-backend:2.8 ./todo-backend
docker push fjpiedade/todo-backend:2.8
```

The frontend continues to use the existing `fjpiedade/todo-app:2.6` image because its application code has not changed.

### 6. Deploy the backend and frontend

```bash
kubectl apply -f manifests/todo-backend-service.yaml
kubectl apply -f manifests/todo-backend-deployment.yaml
kubectl apply -f manifests/todo-app-service.yaml
kubectl apply -f manifests/todo-app-deployment.yaml
```

Wait for both Deployments:

```bash
kubectl rollout status deployment/todo-backend -n project --timeout=120s
kubectl rollout status deployment/todo-app -n project --timeout=120s
```

### 7. Apply Ingress

```bash
kubectl apply -f manifests/ingress.yaml
```

### 8. Check the resources

```bash
kubectl get pods -n project
kubectl get deployments -n project
kubectl get statefulsets -n project
kubectl get services -n project
kubectl get configmaps -n project
kubectl get secrets -n project
kubectl get ingress -n project
kubectl get pvc -n project
kubectl get pv
```

The frontend and backend Deployments should become ready, the PostgreSQL StatefulSet should have one ready replica, and both PersistentVolumeClaims should be `Bound`.

## Inter-pod Communication

The frontend connects to the backend through the Kubernetes Service DNS name:

```text
http://todo-backend-svc:80
```

The backend connects to PostgreSQL using:

```text
todo-postgres:5432
```

Both connections use Kubernetes DNS resolution within the `project` namespace.

The backend receives its database connection settings through environment variables populated from the ConfigMap and Secret.

## Testing

The application can be tested through the configured Ingress.

### Test the application

Retrieve the application HTML:

```bash
curl -i http://localhost:8081/
```

Retrieve todos:

```bash
curl -i http://localhost:8081/todos
```

Create a todo:

```bash
curl -i -X POST http://localhost:8081/todos \
  -H "Content-Type: application/json" \
  -d '{"text":"Todo created through Ingress"}'
```

Retrieve the todos again:

```bash
curl -i http://localhost:8081/todos
```

The POST request should return HTTP `201 Created`, and the new todo should appear in the subsequent GET response.

### Verify database persistence

Query the database directly:

```bash
kubectl exec -n project todo-postgres-0 -- sh -c \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT * FROM todos ORDER BY id;"'
```

The query should return the todos created through the application.

### Test persistence after backend restart

Restart the backend Deployment:

```bash
kubectl rollout restart deployment/todo-backend -n project
kubectl rollout status deployment/todo-backend -n project
```

Retrieve todos again:

```bash
curl -i http://localhost:8081/todos
```

Previously created todos should still be available.

### Test persistence after PostgreSQL Pod restart

Restart the StatefulSet:

```bash
kubectl rollout restart statefulset/todo-postgres -n project
kubectl rollout status statefulset/todo-postgres -n project --timeout=180s
```

Query the database again:

```bash
kubectl exec -n project todo-postgres-0 -- sh -c \
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT * FROM todos ORDER BY id;"'
```

Previously stored todos should remain available after the PostgreSQL Pod has restarted.

### Check application logs

```bash
kubectl logs deployment/todo-app -n project
kubectl logs deployment/todo-backend -n project
kubectl logs statefulset/todo-postgres -n project
```

## Image Caching and Persistence

The application downloads a random image from [Lorem Picsum](https://picsum.photos/) and stores it in persistent storage.

The following settings are configurable through environment variables:

- `IMAGE_FILE`: image file location.
- `IMAGE_SOURCE_URL`: external image source.
- `IMAGE_CACHE_DURATION`: cache duration.
- `IMAGE_HTTP_TIMEOUT`: HTTP client timeout.

The frontend PersistentVolume and PersistentVolumeClaim provide storage that survives application Pod restarts.

This storage is independent of the PostgreSQL storage used for todos.

## Project Structure

```text
exercises/2.8/
├── manifests/
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── postgres.yaml
│   ├── todo-app-deployment.yaml
│   ├── todo-app-service.yaml
│   ├── todo-backend-deployment.yaml
│   └── todo-backend-service.yaml
├── persistent-volume/
│   ├── persistentvolume.yaml
│   └── persistentvolumeclaim.yaml
├── todo-app/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── image/
│   └── static/
├── todo-backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── namespace.yaml
└── README.md
```

## Cleanup

To remove the application resources:

```bash
kubectl delete -f manifests/
kubectl delete -f persistent-volume/
```

The PostgreSQL StatefulSet uses a dynamically provisioned PersistentVolumeClaim. Check the PVC and PV lifecycle before deleting storage if the database data needs to be retained.

To remove the namespace after removing its namespaced resources:

```bash
kubectl delete namespace project
```

Only remove the namespace when it is no longer needed.

## Key Concepts

- Kubernetes ConfigMaps
- Kubernetes Secrets
- Environment variables in containers
- Externalized application configuration
- Docker image builds and publishing
- Kubernetes Namespaces
- Deployments and Services
- StatefulSets
- Headless Services
- PostgreSQL
- PersistentVolumes and PersistentVolumeClaims
- Database connectivity from Kubernetes Pods
- Data persistence across Pod restarts
- Kubernetes internal DNS
- Ingress

## Summary

Exercise 2.8 extends the Todo application by replacing the backend's in-memory storage with PostgreSQL.

The database is deployed as a StatefulSet with one replica and uses persistent storage through a PersistentVolumeClaim.

The backend obtains its database connection settings from a ConfigMap and credentials from a Secret.

Todos are stored in PostgreSQL and remain available after backend and database Pod restarts.

The frontend continues to serve the web interface and cached image, while the backend handles Todo API requests and database operations.
