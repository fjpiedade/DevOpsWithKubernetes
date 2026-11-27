# Exercise 3.10 – Automated Database Backups to Google Cloud Storage

## Objective

Extend the Todo application deployment with an automated backup solution for its PostgreSQL database.

The exercise requires a Kubernetes CronJob that:

- Runs once every 24 hours.
- Creates a database backup using `pg_dump`.
- Uploads the backup to Google Cloud Storage.
- Authenticates securely without storing service account JSON keys.

The solution uses GKE Workload Identity Federation to grant the backup workload access to Google Cloud Storage.

## Application Architecture

The backup solution runs alongside the existing Todo application in the Kubernetes namespace `project`.

It includes:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the application REST API.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Kubernetes Service** – provides internal access to PostgreSQL.
- **Kubernetes Secret** – provides PostgreSQL credentials.
- **Backup ServiceAccount** – identifies the backup workload.
- **Backup CronJob** – schedules and executes database backups.
- **Google Cloud Storage bucket** – stores the generated backup files.

The backup process runs independently of the application containers.

## Implementation

### Backup Docker image

A dedicated Docker image is used to provide the tools required by the backup process:

- PostgreSQL 16 client utilities, including `pg_dump`.
- Google Cloud CLI, including `gcloud storage`.

The Dockerfile is located in:

`exercises/3.10/todo-backup/Dockerfile`

Build the image from the repository root:

```bash
docker build --platform linux/amd64 \
  -t todo-backup:3.10 \
  exercises/3.10/todo-backup
```

Tag the image for Google Artifact Registry:

```bash
docker tag todo-backup:3.10 \
  europe-north1-docker.pkg.dev/dwk-gke-fjp-2026/my-repository/todo-backup:3.10
```

Push the image:

```bash
docker push \
  europe-north1-docker.pkg.dev/dwk-gke-fjp-2026/my-repository/todo-backup:3.10
```

### Google Cloud Storage bucket

A dedicated bucket was created for the database backups:

`gs://dwk-gke-fjp-2026-todo-backups`

The bucket configuration is:

| Setting               | Value                       |
| --------------------- | --------------------------- |
| Location              | `europe-north1`             |
| Storage class         | `STANDARD`                  |
| Access control        | Uniform bucket-level access |
| Public access         | Prevented                   |
| Soft delete retention | 7 days                      |

Backup files are stored under the following prefix:

`gs://dwk-gke-fjp-2026-todo-backups/postgres/`

Each backup filename includes a UTC timestamp.

### Google Cloud service account

A dedicated Google Cloud service account was created:

`todo-backup-sa@dwk-gke-fjp-2026.iam.gserviceaccount.com`

The service account has the following bucket-level IAM roles:

- `roles/storage.objectCreator` – allows the creation of backup objects.
- `roles/storage.objectViewer` – allows the backup process to inspect uploaded objects.

These permissions are granted at bucket level rather than across the entire Google Cloud project.

### GKE Workload Identity Federation

GKE Workload Identity Federation allows Kubernetes workloads to authenticate to Google Cloud using their Kubernetes ServiceAccounts.

The cluster workload identity pool is:

`dwk-gke-fjp-2026.svc.id.goog`

A Kubernetes ServiceAccount named `todo-backup` was created in the `project` namespace.

The ServiceAccount is linked to the Google Cloud service account using the annotation:

`iam.gke.io/gcp-service-account: todo-backup-sa@dwk-gke-fjp-2026.iam.gserviceaccount.com`

The Google Cloud service account also grants the Kubernetes ServiceAccount the `roles/iam.workloadIdentityUser` role.

This configuration allows the backup Pod to access Google Cloud Storage without storing service account keys or credentials in Kubernetes Secrets.

### Kubernetes backup ServiceAccount

The ServiceAccount manifest is located at:

`exercises/3.10/manifests/todo-backup-serviceaccount.yaml`

It defines the identity used by the backup CronJob.

### Kubernetes CronJob

The CronJob manifest is located at:

`exercises/3.10/manifests/todo-backup-cronjob.yaml`

The CronJob is configured with:

| Setting                | Value         |
| ---------------------- | ------------- |
| Name                   | `todo-backup` |
| Namespace              | `project`     |
| Schedule               | `0 0 * * *`   |
| Time zone              | `Etc/UTC`     |
| Concurrency policy     | `Forbid`      |
| Successful Job history | 3             |
| Failed Job history     | 3             |

The schedule runs the backup every day at midnight UTC.

The CronJob uses the `todo-backup` Kubernetes ServiceAccount and the dedicated Docker image stored in Artifact Registry.

Database connection settings are provided through environment variables. The PostgreSQL password is read from the existing `todo-postgres-secret` Kubernetes Secret.

The database is accessed through its Kubernetes DNS name:

`todo-postgres.project.svc.cluster.local`

### Backup process

Each CronJob execution performs the following operations:

1. Generates a UTC timestamp for the backup filename.
2. Connects to the PostgreSQL database.
3. Creates a custom-format backup using `pg_dump`.
4. Uploads the backup to the Google Cloud Storage bucket.
5. Reports successful completion in the container logs.

The backup command uses:

```bash
pg_dump --format=custom --no-owner --no-acl \
  --file="${BACKUP_FILE}"
```

Custom-format backups can be restored using PostgreSQL's `pg_restore` utility.

The resulting files follow this naming convention:

`todos-YYYYMMDDTHHMMSSZ.dump`

## Deployment

All Kubernetes resources are maintained in:

`exercises/3.10/`

The Kustomize configuration includes the backup ServiceAccount and CronJob manifests.

Render the Kubernetes resources:

```bash
kubectl kustomize exercises/3.10
```

Validate the manifests against the cluster API:

```bash
kubectl apply --dry-run=server -k exercises/3.10
```

Apply the backup resources:

```bash
kubectl apply -f exercises/3.10/manifests/todo-backup-serviceaccount.yaml
kubectl apply -f exercises/3.10/manifests/todo-backup-cronjob.yaml
```

The application and PostgreSQL resources are expected to be running in the `project` namespace.

## Testing

A manual Job can be created from the CronJob to test the backup process without waiting for its scheduled execution:

```bash
kubectl create job todo-backup-test \
  --from=cronjob/todo-backup \
  -n project
```

Inspect the Job:

```bash
kubectl get jobs -n project
```

Inspect its logs:

```bash
kubectl logs job/todo-backup-test -n project
```

A successful execution should show messages similar to:

```text
Creating PostgreSQL backup...
Uploading backup to Google Cloud Storage...
Copying file:///tmp/todos-...dump to gs://dwk-gke-fjp-2026-todo-backups/postgres/todos-...dump
Backup completed successfully.
```

### Verify uploaded backups

List the objects stored in the bucket:

```bash
gcloud storage ls -l \
  gs://dwk-gke-fjp-2026-todo-backups/postgres/
```

The backup process was successfully tested.

Verified backup:

- Object: `todos-20261127T201952Z.dump`
- Size: 3,163 bytes.

The manual test Jobs were removed after verification. The scheduled CronJob remains configured in the cluster.

## Restoring a Backup

A backup can be downloaded from Google Cloud Storage:

```bash
gcloud storage cp \
  gs://dwk-gke-fjp-2026-todo-backups/postgres/todos-20261127T201952Z.dump \
  ./todos-backup.dump
```

Inspect the backup contents:

```bash
pg_restore --list ./todos-backup.dump
```

Restore the backup into an existing PostgreSQL database:

```bash
pg_restore --no-owner --no-acl \
  --dbname=todos \
  ./todos-backup.dump
```

The restore operation should be performed against the intended recovery database, with the appropriate connection credentials and database state.

## Validation Results

The following tests and checks were successfully performed:

- The dedicated backup Docker image was built successfully.
- The image was pushed to Google Artifact Registry.
- The Google Cloud Storage bucket was created with private access.
- The Google Cloud service account was configured with bucket-level permissions.
- GKE Workload Identity Federation was configured and verified.
- The Kubernetes backup ServiceAccount was created and linked to the Google Cloud identity.
- The backup CronJob was deployed successfully.
- A manual Job successfully connected to PostgreSQL and created a custom-format backup.
- The backup file was successfully uploaded to Google Cloud Storage.
- The uploaded object was verified using `gcloud storage ls`.
- Temporary test Jobs and diagnostic Pods were removed.

## Result

Exercise 3.10 provides an automated PostgreSQL backup solution for the Todo application.

The Kubernetes CronJob runs once every 24 hours, creates a database backup using `pg_dump` and stores it in Google Cloud Storage.

GKE Workload Identity Federation provides secure authentication without requiring service account JSON keys.

The backup process is defined declaratively through Kubernetes manifests and integrated into the exercise's Kustomize configuration.
