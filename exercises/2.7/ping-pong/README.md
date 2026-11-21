# Ping-pong — PostgreSQL Integration

## Overview

The Ping-pong application is a Go HTTP server that stores its counter in a PostgreSQL database instead of a local file.

This implementation ensures that the counter persists across application Pod restarts and PostgreSQL Pod restarts, provided the database's PersistentVolumeClaim is retained.

## HTTP endpoints

| Method | Endpoint    | Description                                          |
| ------ | ----------- | ---------------------------------------------------- |
| GET    | `/pingpong` | Increments the counter and returns `pong N`.         |
| GET    | `/pings`    | Returns the current counter without incrementing it. |

## PostgreSQL connection

The application connects to PostgreSQL using the following environment variables:

| Variable      | Description                 |
| ------------- | --------------------------- |
| `DB_HOST`     | PostgreSQL Service hostname |
| `DB_PORT`     | PostgreSQL port             |
| `DB_USER`     | Database username           |
| `DB_PASSWORD` | Database password           |
| `DB_NAME`     | Database name               |

Database credentials are provided through a Kubernetes Secret.

The application uses the `github.com/lib/pq` PostgreSQL driver.

## Database schema

At startup, the application creates the counter table if it does not exist:

```sql
CREATE TABLE IF NOT EXISTS pingpong_counter (
    id INTEGER PRIMARY KEY,
    count BIGINT NOT NULL
);
```

The initial counter row is inserted only if it does not already exist:

```sql
INSERT INTO pingpong_counter (id, count)
VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;
```

The counter is incremented atomically:

```sql
UPDATE pingpong_counter
SET count = count + 1
WHERE id = 1
RETURNING count;
```

## Container image

The application image is:

`fjpiedade/ping-pong:2.7`

The Dockerfile uses a multi-stage build to compile the Go application and produce a lightweight runtime image.

## Local development

Download dependencies:

```bash
go mod tidy
```

Build the application:

```bash
go build -o ping-pong .
```

Run tests:

```bash
go test ./...
```

The application requires a reachable PostgreSQL database to start.

## Kubernetes deployment

The Ping-pong Deployment connects to the PostgreSQL StatefulSet through the internal Service:

`postgres:5432`

The application is exposed internally through the `ping-pong-svc` ClusterIP Service and externally through the Kubernetes Ingress.

## Validation

Increment the Ping-pong counter through the Ingress:

```bash
curl http://localhost:8081/pingpong
```

Access the Log-output application through the Ingress:

```bash
curl http://localhost:8081/
```

Access the Ping-pong counter directly through its Kubernetes Service:

```bash
kubectl run ping-test --rm -i --restart=Never \
  --image=curlimages/curl:8.12.1 \
  -n exercises \
  -- curl -sS http://ping-pong-svc/pings
```

Inspect the database directly:

```bash
kubectl exec -it postgres-0 -n exercises -- sh
psql -U pingpong -d pingpong
```

Then execute:

```sql
SELECT * FROM pingpong_counter;
```

The counter can also be incremented directly in PostgreSQL:

```sql
UPDATE pingpong_counter
SET count = count + 1
WHERE id = 1
RETURNING count;
```

## Key concepts

- Go database connectivity.
- PostgreSQL.
- SQL table initialization.
- Atomic counter updates.
- Kubernetes Secrets.
- StatefulSets and PersistentVolumeClaims.
- Persistent application data.
