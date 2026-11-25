# Ping-pong — PostgreSQL Integration

## Overview

The Ping-pong application is a Go HTTP server that stores its counter in a PostgreSQL database instead of a local file.

The counter persists across application Pod restarts and PostgreSQL Pod restarts, provided the database's PersistentVolumeClaim is retained.

## HTTP endpoints

| Method | Endpoint    | Description                                                           |
| ------ | ----------- | --------------------------------------------------------------------- |
| GET    | `/`         | Health check endpoint. Returns `200 OK` without changing the counter. |
| GET    | `/pingpong` | Increments the counter and returns `pong N`.                          |
| GET    | `/pings`    | Returns the current counter without incrementing it.                  |

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

## Kubernetes resources

- `ping-pong` Deployment running the application.
- `ping-pong-svc` ClusterIP Service, exposing the application internally on port 80 and forwarding to port 9901.
- `postgres` StatefulSet and headless Service.
- `postgres-secret` Secret containing database credentials.
- PersistentVolumeClaim for PostgreSQL data storage.
- Gateway API `Gateway` and `HTTPRoute` for external HTTP access.

## Gateway API routing

The GKE Gateway routes requests to the Ping-pong Service:

| Path        | Backend         | Description                        |
| ----------- | --------------- | ---------------------------------- |
| `/pingpong` | `ping-pong-svc` | Increments and returns the counter |
| `/pings`    | `ping-pong-svc` | Returns the current counter        |

The `/` path is routed to the Log-output Service.

The Log-reader also accesses the current counter internally through the Kubernetes Service DNS name `ping-pong-svc`, using the `/pings` endpoint.

The Ping-pong `/` endpoint remains available for Kubernetes readiness and liveness probes. It returns `200 OK` without modifying the counter.
