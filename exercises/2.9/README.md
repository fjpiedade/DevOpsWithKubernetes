# Exercise 2.9 – Generate todos with a Kubernetes CronJob

## Objective

Extend the Todo application developed in Exercise 2.8 by introducing a Kubernetes CronJob that automatically generates a new Todo every hour.

Each generated Todo contains a randomly selected Wikipedia article URL in the format:

`Read <URL>`

## Dependency: Exercise 2.8

**Exercise 2.9 depends on Exercise 2.8 and is not a standalone deployment.**

The complete application infrastructure from Exercise 2.8 must be deployed and running in the Kubernetes cluster before the CronJob can work.

Exercise 2.8 provides the application and infrastructure required by the generator:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the REST API used to create and retrieve Todos.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Backend and frontend Services** – provide internal communication between application components.
- **PostgreSQL headless Service** – provides stable network identity for the database.
- **ConfigMap** – provides backend database connection settings.
- **Secret** – stores PostgreSQL credentials.
- **PersistentVolumeClaims and PersistentVolumes** – provide persistent storage for the application and database.
- **Ingress** – exposes the Todo application externally.
- **Namespace `project`** – contains the application resources.

The CronJob communicates with the existing backend through its Kubernetes Service:

`http://todo-backend-svc:80`

It does not connect directly to PostgreSQL and does not require database credentials.

### Required deployment order

1. Deploy all resources from Exercise 2.8.
2. Verify that the frontend, backend and PostgreSQL Pods are running.
3. Verify that the backend Service is available.
4. Deploy the Exercise 2.9 CronJob.

If the Exercise 2.8 backend or its dependencies are unavailable, the generator cannot create new Todos.

## Implementation

### Todo Generator

A standalone Go application was created in `todo-generator/`.

The application:

1. Requests `https://en.wikipedia.org/wiki/Special:Random`.
2. Follows the HTTP redirect to obtain the final Wikipedia article URL.
3. Creates a Todo with the text `Read <URL>`.
4. Sends the Todo to the existing backend using an HTTP POST request.

The application uses Go's standard library and has no external dependencies.

### Docker Image

The generator was packaged into a Docker image:

- Image: `fjpiedade/todo-generator:2.9`

### Kubernetes CronJob

The CronJob is defined in `manifests/todo-generator-cronjob.yaml`.

Configuration:

- Namespace: `project`
- Schedule: `0 * * * *`
- Time zone: `Etc/UTC`
- Concurrency policy: `Forbid`
- Backend URL: `http://todo-backend-svc:80`
- Successful Job history limit: 3
- Failed Job history limit: 3

The CronJob runs at the beginning of every hour and creates a new Todo using the existing backend.

## Deployment

**Prerequisite:** Exercise 2.8 must already be deployed and running.

Verify the application resources:

```bash
kubectl get deployments,statefulsets,pods,services -n project
kubectl get configmaps,secrets,pvc,ingress -n project
```

Apply the CronJob:

```bash
kubectl apply -f exercises/2.9/manifests/todo-generator-cronjob.yaml
```

Inspect it:

```bash
kubectl get cronjob -n project
kubectl describe cronjob todo-generator -n project
```

## Testing

A manual Job was created from the CronJob to validate its execution without waiting for the next scheduled hour:

```bash
kubectl create job \
  --from=cronjob/todo-generator \
  todo-generator-manual-$(date +%s) \
  -n project
```

Inspect the Job and its logs:

```bash
kubectl get jobs,pods -n project
kubectl logs -n project job/<job-name>
```

Verify the generated Todo through the backend:

```bash
curl http://localhost:8080/todos
```

The `localhost:8080` endpoint in this test requires a port-forward to the backend Service:

```bash
kubectl port-forward -n project service/todo-backend-svc 8080:80
```

## Validation Results

- Confirmed that the Exercise 2.8 application infrastructure was running.
- Successfully retrieved a random Wikipedia article.
- Successfully created a Todo through the backend Kubernetes Service.
- Manual Job completed successfully (`1/1`).
- Generator Pod completed without restarts.
- Confirmed the generated Todo was returned by the backend.
- Confirmed the Todo was stored in the existing PostgreSQL-backed application.

## Result

Exercise 2.9 extends the persistent Todo application from Exercise 2.8 with automated Todo generation.

The existing frontend, backend and PostgreSQL infrastructure remain unchanged. The new CronJob periodically creates Todos containing randomly selected Wikipedia article URLs, which are stored persistently in PostgreSQL through the existing backend.
