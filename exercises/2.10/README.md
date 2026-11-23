# Exercise 2.10 – Logging and Todo Validation

## Objective

Extend the Todo application with backend request logging and server-side validation.

This exercise is self-contained: all application source code and Kubernetes manifests required to deploy the application are included in this directory.

The exercise introduces:

- HTTP request logging in the Todo backend.
- Logging of received, accepted and rejected Todos.
- A maximum Todo text length of 140 characters, enforced by the backend.
- Centralized log collection and querying with Alloy, Loki and Grafana.

## Application Architecture

The application is deployed in the Kubernetes namespace `project` and includes:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the REST API and validates Todo text.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Kubernetes Services** – provide internal communication between application components.
- **ConfigMaps and Secret** – provide application configuration and database credentials.
- **PersistentVolumes and PersistentVolumeClaims** – provide persistent storage.
- **Ingress** – exposes the Todo application externally.

All application source code, Dockerfiles and Kubernetes manifests are maintained in this exercise directory.

### Monitoring infrastructure

The application logs are collected and visualized using:

- **Alloy** – collects Kubernetes container logs.
- **Loki** – stores and indexes the logs.
- **Grafana** – provides the interface for querying and inspecting logs.
- **Prometheus** – collects Kubernetes and application infrastructure metrics.

These monitoring components are infrastructure dependencies; they are installed separately from the Todo application.

## Implementation

### Backend request logging

The backend logs HTTP requests, including:

- HTTP method and request path.
- HTTP response status code.
- Request duration.

The backend also logs Todo processing events:

- Todo received, including its text and character length.
- Todo accepted, including its database ID.
- Todo rejected when its text exceeds the permitted length.

The request logging middleware records the actual HTTP response status, including error responses.

### Todo text validation

The backend enforces a maximum text length of 140 characters.

- Texts containing up to 140 characters are accepted.
- Texts containing more than 140 characters are rejected.
- Rejected requests return HTTP `400 Bad Request`.
- Successfully created Todos return HTTP `201 Created`.

Validation is performed by the backend, independently of frontend validation.

### Docker images

The backend image for this exercise:

- `fjpiedade/todo-backend:2.10`

The frontend continues to use:

- `fjpiedade/todo-app:2.6`

Build and publish the backend image from the repository root:

```bash
docker build -t fjpiedade/todo-backend:2.10 ./exercises/2.10/todo-backend
docker push fjpiedade/todo-backend:2.10
```

The Kubernetes backend Deployment uses this image with `imagePullPolicy: Always`.

## Monitoring Stack Installation

The monitoring infrastructure is installed using Helm.

All configuration files required by this exercise are included in:

`exercises/2.10/helm/`

### Prerequisites

- A running Kubernetes cluster.
- `kubectl` configured to access the cluster.
- Helm installed.

### Helm configuration files

| File                          | Purpose                                                |
| ----------------------------- | ------------------------------------------------------ |
| `prom-values.yaml`            | Prometheus configuration                               |
| `loki-values.yaml`            | Loki single-binary configuration                       |
| `k8smon-values.yaml`          | Kubernetes Monitoring and log collection configuration |
| `alloy-singleton-values.yaml` | Alloy Singleton configuration                          |
| `grafana-values.yaml`         | Grafana credentials and data sources                   |

### Add Helm repositories

Run these commands from the repository root:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

### Install Prometheus

```bash
helm upgrade --install prom prometheus-community/prometheus \
  --namespace monitoring \
  --create-namespace \
  --version 29.49.0 \
  --values exercises/2.10/helm/prom-values.yaml
```

### Install Loki

```bash
helm upgrade --install loki grafana/loki \
  --namespace monitoring \
  --version 7.7.0 \
  --values exercises/2.10/helm/loki-values.yaml
```

Loki is configured in single-binary mode with filesystem storage, suitable for this local Kubernetes exercise environment.

### Install Kubernetes Monitoring

```bash
helm upgrade --install k8smon grafana/k8s-monitoring \
  --namespace monitoring \
  --version 4.6.1 \
  --values exercises/2.10/helm/k8smon-values.yaml
```

This chart configures Kubernetes monitoring and Pod log collection. The configuration sends logs to the Loki gateway using tenant ID `1`.

### Install Alloy Singleton

```bash
helm upgrade --install k8smon-alloy-singleton grafana/alloy \
  --namespace monitoring \
  --version 1.14.4 \
  --values exercises/2.10/helm/alloy-singleton-values.yaml
```

This release provides the Alloy Singleton collector used to read container log files and forward them to Loki.

### Install Grafana

```bash
helm upgrade --install grafana grafana/grafana \
  --namespace monitoring \
  --version 10.5.15 \
  --values exercises/2.10/helm/grafana-values.yaml
```

Grafana is configured with two data sources:

- Prometheus: `http://prom-prometheus-server`
- Loki: `http://loki-gateway.monitoring.svc.cluster.local`

The Loki data source includes the required `X-Scope-OrgID` HTTP header with tenant ID `1`.

### Verify the monitoring stack

Check the Helm releases:

```bash
helm list -n monitoring
```

Check the monitoring Pods:

```bash
kubectl get pods -n monitoring
```

The main monitoring components should reach the `Running` state.

### Access Grafana

Forward the Grafana Service to localhost:

```bash
kubectl port-forward --namespace monitoring svc/grafana 3000:80
```

Open:

http://localhost:3000

The credentials configured in `grafana-values.yaml` are:

- Username: `admin`
- Password: `admin`

Navigate to **Explore** and select Loki to query application logs.

## Deployment

All application manifests are included in this directory.

Apply the resources from the repository root:

```bash
kubectl apply -f exercises/2.10/namespace.yaml
kubectl apply -f exercises/2.10/persistent-volume/
kubectl apply -f exercises/2.10/manifests/
```

Verify the application:

```bash
kubectl get deployments,statefulsets,pods,services -n project
kubectl get configmaps,secrets,pvc,ingress -n project
```

The backend Deployment must use `fjpiedade/todo-backend:2.10`.

If the backend image has been updated, restart the Deployment and wait for the rollout:

```bash
kubectl rollout restart deployment/todo-backend -n project
kubectl rollout status deployment/todo-backend -n project --timeout=120s
```

## Testing

Forward the backend Service to localhost:

```bash
kubectl port-forward -n project service/todo-backend-svc 8080:80
```

Keep the port-forward running in a separate terminal.

### Test a valid Todo (140 characters)

```bash
curl -i -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d "{\"text\":\"$(printf 'A%.0s' {1..140})\"}"
```

Expected result:

- HTTP `201 Created`.
- The Todo is stored successfully.
- The backend logs the received and accepted Todo.

### Test an invalid Todo (141 characters)

```bash
curl -i -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d "{\"text\":\"$(printf 'A%.0s' {1..141})\"}"
```

Expected result:

- HTTP `400 Bad Request`.
- Response: `Todo text cannot exceed 140 characters`.
- The backend logs the received Todo and its rejection.

### Inspect backend logs

```bash
kubectl logs deployment/todo-backend -n project --since=15m
```

The logs should contain entries similar to:

```text
Todo received: text="..." | length=141
Todo rejected: text exceeds 140 characters | length=141
POST /todos | status=400 | duration=...

Todo received: text="..." | length=140
Todo accepted: id=... | text="..."
POST /todos | status=201 | duration=...
```

## Monitoring Logs with Grafana and Loki

Alloy collects the backend container logs and forwards them to Loki.

Open Grafana and navigate to **Explore**. Select the Loki data source and use the following LogQL queries.

Show logs from the project namespace:

```logql
{namespace="project"}
```

Show rejected Todos:

```logql
{namespace="project"} |= "Todo rejected"
```

Show HTTP POST requests to the Todo endpoint:

```logql
{namespace="project"} |= "POST /todos"
```

Filter specifically for rejected Todos from the backend container:

```logql
{namespace="project", container="todo-backend"} |= "Todo rejected"
```

Use a suitable time range, such as **Last 1 hour**, to inspect recent events.

## Validation Results

The following tests were successfully performed:

- Backend Deployment successfully rolled out with image `fjpiedade/todo-backend:2.10`.
- A Todo containing 140 characters was accepted with HTTP `201 Created`.
- A Todo containing 141 characters was rejected with HTTP `400 Bad Request`.
- Backend logs recorded received, accepted and rejected Todos.
- HTTP request logs recorded the correct response status codes (`201` and `400`).
- Alloy collected the backend logs from the `project` namespace.
- Loki stored the logs and Grafana successfully displayed them.
- The rejected Todo was visible in Grafana using the query `{namespace="project"} |= "Todo rejected"`.
- HTTP POST request logs were visible in Grafana using `{namespace="project"} |= "POST /todos"`.

## Result

Exercise 2.10 provides a self-contained Todo application deployment with backend request logging and server-side Todo text validation.

The backend enforces the 140-character limit and records successful and rejected requests. Alloy, Loki and Grafana provide centralized collection, storage and inspection of application logs.
