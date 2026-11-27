# Exercise 3.6 – The Project, Step 15: Automatic Deployment to Google Kubernetes Engine

## Objective

Configure automatic deployment of the Todo application project to Google Kubernetes Engine (GKE) using GitHub Actions.

This exercise builds on Exercise 3.5, extending the Kustomize-based deployment with a continuous integration and continuous deployment (CI/CD) workflow.

The main objectives are:

- Automate Docker image builds using GitHub Actions.
- Publish application images to Google Artifact Registry.
- Authenticate GitHub Actions securely using Workload Identity Federation.
- Automatically deploy application updates to GKE using Kustomize.
- Verify successful application rollouts.
- Avoid persistent volume conflicts during application updates.

## Application Architecture

The application is deployed in the Kubernetes namespace `project` and includes:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the REST API and processes Todo requests.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Todo generator CronJob** – periodically creates Todo entries.
- **Kubernetes Services** – provide internal communication between application components.
- **ConfigMaps and Secrets** – provide application configuration and database credentials.
- **PersistentVolumeClaims** – provide persistent storage.
- **Ingress** – exposes the Todo application externally.

The application source code, Dockerfiles and Kubernetes manifests are maintained in `exercises/3.6/`.

## CI/CD Architecture

The deployment process is automated through GitHub Actions.

Whenever a commit is pushed to the `main` branch, the workflow:

1. Checks out the repository.
2. Authenticates with Google Cloud using Workload Identity Federation.
3. Configures Docker authentication for Artifact Registry.
4. Retrieves credentials for the GKE cluster.
5. Builds the three application Docker images.
6. Pushes the images to Google Artifact Registry.
7. Updates the image references in the Kustomize configuration.
8. Applies the Kubernetes resources to GKE.
9. Waits for the frontend and backend Deployments to complete their rollouts.

The workflow is defined in:

```text
.github/workflows/main.yaml
```

## Google Artifact Registry

A Docker repository was created in Google Artifact Registry.

Repository configuration:

| Setting    | Value                          |
| ---------- | ------------------------------ |
| Project ID | `dwk-gke-fjp-2026`             |
| Location   | `europe-north1`                |
| Repository | `my-repository`                |
| Format     | Docker                         |
| Registry   | `europe-north1-docker.pkg.dev` |

The workflow builds and publishes three images:

| Component      | Artifact Registry image |
| -------------- | ----------------------- |
| Todo App       | `todo-app`              |
| Todo Backend   | `todo-backend`          |
| Todo Generator | `todo-generator`        |

Each image is tagged using the `main` branch name and the full Git commit SHA.

For example:

```text
main-COMMIT_SHA
```

This allows each deployment to be associated with the exact source commit that produced the images.

The PostgreSQL image remains an external image and is not rebuilt by this workflow.

## Secure Authentication with Workload Identity Federation

GitHub Actions authenticates with Google Cloud through Workload Identity Federation (WIF).

This approach avoids storing Google Cloud service account JSON keys in GitHub.

The configuration includes:

- A Google Cloud Workload Identity Pool named `github-pool`.
- An OIDC provider named `github-provider`.
- A dedicated service account named `github-artifact-reg`.
- A repository and branch restriction for `fjpiedade/DevOpsWithKubernetes` and `main`.

The service account uses the following IAM roles:

- `roles/artifactregistry.writer` – allows publishing images to the configured Artifact Registry repository.
- `roles/container.clusterViewer` – allows retrieving GKE cluster information and credentials.
- `roles/container.developer` – allows managing Kubernetes resources.

The Workload Identity provider condition restricts authentication to the intended GitHub repository and the `main` branch.

## GitHub Repository Variables

The workflow uses GitHub repository variables to configure the deployment.

These variables are defined under:

**Repository Settings → Secrets and variables → Actions → Variables**

| Variable                     | Value                                                                                                |
| ---------------------------- | ---------------------------------------------------------------------------------------------------- |
| `PROJECT_ID`                 | `dwk-gke-fjp-2026`                                                                                   |
| `GKE_CLUSTER`                | `dwk-cluster`                                                                                        |
| `GKE_ZONE`                   | `europe-north1-b`                                                                                    |
| `REGISTRY`                   | `europe-north1-docker.pkg.dev`                                                                       |
| `REPOSITORY`                 | `my-repository`                                                                                      |
| `WORKLOAD_IDENTITY_PROVIDER` | `projects/284041604326/locations/global/workloadIdentityPools/github-pool/providers/github-provider` |
| `SERVICE_ACCOUNT`            | `github-artifact-reg@dwk-gke-fjp-2026.iam.gserviceaccount.com`                                       |

These values are configuration identifiers, not private credentials.

## GitHub Actions Workflow

The workflow is triggered by pushes to the `main` branch.

The workflow file is:

```text
.github/workflows/main.yaml
```

It uses the following GitHub Actions permissions:

```yaml
permissions:
  contents: read
  id-token: write
```

The `id-token: write` permission allows GitHub Actions to request an OIDC token for Workload Identity Federation.

The workflow builds and pushes the three Docker images and deploys the resulting versions using Kustomize.

The Kustomize image references are updated dynamically during the workflow execution. The generated image references point to the corresponding Artifact Registry images and commit-specific tags.

The workflow modifies only the runner's working copy of the Kustomize configuration. It does not automatically commit those generated image changes to the repository.

## Kustomize Configuration

The project uses a root-level `kustomization.yaml` file to define the Kubernetes resources.

The configuration includes:

- PersistentVolumeClaim for frontend image storage.
- Application ConfigMaps.
- PostgreSQL Secret and StatefulSet.
- Todo frontend Deployment and Service.
- Todo backend Deployment and Service.
- Todo generator CronJob.
- GKE Ingress.

The `project` namespace is created separately and is expected to exist before the workflow applies the application resources.

The resources are managed together through Kustomize.

### Project structure

```text
.github/
└── workflows/
    └── main.yaml

exercises/3.6/
├── kustomization.yaml
├── namespace.yaml
├── manifests/
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── postgres.yaml
│   ├── todo-app-deployment.yaml
│   ├── todo-app-service.yaml
│   ├── todo-backend-deployment.yaml
│   ├── todo-backend-service.yaml
│   └── todo-generator-cronjob.yaml
├── persistent-volume/
│   └── persistentvolumeclaim.yaml
├── helm/
├── todo-app/
│   ├── Dockerfile
│   └── main.go
├── todo-backend/
│   ├── Dockerfile
│   └── main.go
├── todo-generator/
│   ├── Dockerfile
│   └── main.go
└── README.md
```

### Render the Kustomize configuration

From the repository root, inspect the resources generated by Kustomize:

```bash
kubectl kustomize exercises/3.6/
```

This renders the Kubernetes configuration without applying it to the cluster.

## Persistent Volume Update Strategy

The Todo frontend uses a PersistentVolumeClaim with the `ReadWriteOnce` access mode.

The default Kubernetes Deployment strategy is `RollingUpdate`. During an update, this strategy may attempt to create a new Pod before terminating the existing one, potentially causing volume attachment conflicts.

To avoid this problem, the Todo frontend Deployment uses:

```yaml
spec:
  replicas: 1
  strategy:
    type: Recreate
```

The `Recreate` strategy terminates the existing Pod before creating its replacement, preventing simultaneous access attempts by multiple Pods to the same ReadWriteOnce volume.

## Automatic Deployment

### Prerequisites

- A Google Cloud project with billing enabled.
- A running Google Kubernetes Engine cluster.
- An Artifact Registry Docker repository.
- The Workload Identity Federation configuration.
- The GitHub repository variables configured.
- The `project` namespace already created in GKE.
- GitHub Actions enabled for the repository.

### Trigger the deployment

From the repository root, commit and push changes to the `main` branch.

```bash
git add .github/workflows/main.yaml exercises/3.6/ .gitignore
git commit -m "Complete exercise 3.6: automate deployment to GKE"
git push origin main
```

The push automatically triggers the GitHub Actions workflow.

### Monitor the workflow

Open the GitHub repository and navigate to:

**Actions → Build and deploy to GKE**

The workflow should complete all steps successfully.

### Verify the deployed resources

After the workflow completes, check the Deployments:

```bash
kubectl get deployments -n project
```

Check the Pods:

```bash
kubectl get pods -n project
```

Check the CronJob:

```bash
kubectl get cronjob -n project
```

Check the application services:

```bash
kubectl get services -n project
```

Check the Ingress:

```bash
kubectl get ingress -n project
```

The frontend and backend Deployments should complete their rollouts successfully.

## Application Validation

### Verify the deployed images

Inspect the images used by the frontend and backend Deployments:

```bash
kubectl get deployments -n project \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```

The application images should reference the Google Artifact Registry repository and the commit-specific tag.

### Verify the application endpoints

Retrieve the external Ingress IP:

```bash
kubectl get ingress todo-app-ingress -n project
```

Test the frontend:

```bash
curl -i http://INGRESS_EXTERNAL_IP/
```

Expected result: HTTP `200 OK` with the Todo application HTML.

Test the Todo API:

```bash
curl -i http://INGRESS_EXTERNAL_IP/todos
```

Expected result: HTTP `200 OK` with a JSON response containing Todo entries.

Create a Todo using the application interface in a browser and verify that it appears in the Todo list.

## Result

Exercise 3.6 extends the Kustomize-based GKE deployment from Exercise 3.5 with an automated CI/CD pipeline.

GitHub Actions builds and publishes the Todo frontend, backend and generator images to Google Artifact Registry.

Workload Identity Federation provides secure authentication between GitHub Actions and Google Cloud without requiring long-lived service account keys.

The workflow automatically updates the application image references and deploys the Kubernetes resources to GKE using Kustomize.

The Todo frontend uses the `Recreate` Deployment strategy to avoid conflicts with its ReadWriteOnce persistent volume during updates.

## Cleanup

Google Cloud resources such as GKE clusters, persistent disks, Ingress resources and external Load Balancers may incur charges.

After completing the exercise and saving the required repository changes, the application resources managed by Kustomize can be removed with:

The `project` namespace is not included in the Kustomize resources and will therefore remain in the cluster. This is intentional because the namespace was created in Exercise 3.5 and may be shared with other exercises.

```bash
kubectl delete -k exercises/3.6/
```

After the Ingress and its associated Load Balancer resources have been released, delete the GKE cluster if it is no longer required.

Verify that no unused Load Balancer resources or persistent disks remain in the Google Cloud project.
