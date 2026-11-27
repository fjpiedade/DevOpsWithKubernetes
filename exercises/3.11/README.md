# Exercise 3.11 – Kubernetes Resource Requests and Limits

## Objective

Configure sensible CPU and memory requests and limits for the Todo application running on Kubernetes.

The objective is to:

- Define resource requests for each application container.
- Set CPU and memory limits to prevent excessive resource consumption.
- Observe actual resource usage with `kubectl top pods`.
- Compare configured resources with the measured usage.

## Application Architecture

The exercise uses the existing Todo application deployment in the Kubernetes namespace `project`.

It includes:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the application REST API.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Todo generator CronJob** – periodically generates Todo items.
- **Backup CronJob** – creates PostgreSQL backups and uploads them to Google Cloud Storage.

## Resource Configuration

Resource requests and limits were added to all five containers.

Requests represent the minimum CPU and memory resources Kubernetes uses when scheduling Pods.

Limits define the maximum CPU and memory resources a container can consume.

### Configured resources

| Container | CPU request | Memory request | CPU limit | Memory limit |
| --------- | ----------- | -------------- | --------- | ------------ |
| todo-app | 25m | 32Mi | 250m | 128Mi |
| todo-backend | 25m | 32Mi | 250m | 256Mi |
| postgres | 100m | 128Mi | 500m | 512Mi |
| todo-generator | 25m | 32Mi | 250m | 128Mi |
| database-backup | 50m | 128Mi | 500m | 512Mi |

The values were selected based on the lightweight workload and the observed resource consumption.

The PostgreSQL PersistentVolumeClaim retains its separate storage request of `1Gi`. This is persistent storage, not container memory or CPU.

### Kubernetes manifests

Resource configuration is defined directly in the container specifications:

- `manifests/todo-app-deployment.yaml`
- `manifests/todo-backend-deployment.yaml`
- `manifests/postgres.yaml`
- `manifests/todo-generator-cronjob.yaml`
- `manifests/todo-backup-cronjob.yaml`

Each container uses the Kubernetes `resources` field with `requests` and `limits`.

## Deployment

Render the Kubernetes resources:

```bash
kubectl kustomize exercises/3.11
```

Validate the manifests against the cluster API:

```
kubectl apply --dry-run=server -k exercises/3.11
```

Apply the configuration:

```
kubectl apply -k exercises/3.11
```

Check the Pods:

```
kubectl get pods -n project
```

The frontend, backend and PostgreSQL Pods should reach the `Running` state. Completed Todo generator Jobs are expected to show `Completed`.

## **Measuring Resource Usage**

Kubernetes Metrics Server provides actual CPU and memory usage.

Inspect current Pod resource consumption:

```
kubectl top pods -n project
```

Inspect the configured requests and limits:

```
kubectl describe pods -n project
```

The `Limits` and `Requests` sections show the effective resource configuration for each running container.

## **Testing and Results**

The manifests were validated successfully using the Kubernetes API server dry-run.

The resources were then applied to the GKE cluster.

The application Pods restarted with the new configuration and reached the `Running` state:

- Todo frontend: `Running`
- Todo backend: `Running`
- PostgreSQL: `Running`

The Todo generator Jobs completed successfully.

### **Observed resource usage**

The following measurements were obtained using `kubectl top pods -n project`:

| Pod           | CPU usage | Memory usage |
| ------------- | --------- | ------------ |
| todo-app      | 1m        | 1Mi          |
| todo-backend  | 0m        | 1Mi          |
| todo-postgres | 11m       | 22Mi         |

The measured usage was comfortably below the configured requests and limits.

The resource configuration was also verified using:

```
kubectl describe pods -n project
```

The output confirmed the expected CPU and memory requests and limits for the frontend, backend and PostgreSQL containers.

The CronJob containers also have resource requests and limits defined. They are not included in the `kubectl top pods` measurements because no corresponding Job was running at the time of measurement.

## **Result**

Exercise 3.11 configures CPU and memory requests and limits for all Todo application containers and scheduled workloads.

The configuration was successfully validated and deployed to GKE.

Resource usage was measured using `kubectl top pods` and confirmed to be well below the configured limits.

This provides Kubernetes with resource requirements for scheduling and helps prevent individual containers from consuming excessive CPU or memory.
