# Ping-pong — PostgreSQL Integration

## Overview

The Ping-pong application is a Go HTTP server that stores its counter in a PostgreSQL database instead of a local file.

The counter persists across application Pod restarts and PostgreSQL Pod restarts, provided the database's PersistentVolumeClaim is retained.

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
