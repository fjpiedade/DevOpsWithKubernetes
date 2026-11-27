# Exercise 3.7 – The Project, Step 16: Separate Environments for Each Branch

## Objective

Improve the automatic deployment pipeline from Exercise 3.6 so that each Git branch is deployed to an independent Kubernetes namespace in Google Kubernetes Engine (GKE).

This exercise extends the existing CI/CD workflow to support multiple deployment environments.

The main objectives are:

- Trigger automatic deployments on pushes to any branch.
- Keep the `main` branch deployed in the `project` namespace.
- Deploy other branches to namespaces matching their branch names.
- Automatically create namespaces when they do not exist.
- Use Kustomize to assign the appropriate namespace to all application resources.
- Build and publish application images to Google Artifact Registry.
- Deploy and verify each environment independently.

## Application Architecture

Each environment contains its own Kubernetes resources:

- **Todo frontend** – provides the user interface.
- **Todo backend** – exposes the REST API and processes Todo requests.
- **PostgreSQL StatefulSet** – stores Todo records persistently.
- **Todo generator CronJob** – periodically creates Todo entries.
- **Kubernetes Services** – provide internal communication between application components.
- **ConfigMaps and Secrets** – provide application configuration and database credentials.
- **PersistentVolumeClaims** – provide persistent storage.
- **Ingress** – exposes the Todo application externally.

The application source code, Dockerfiles and Kubernetes manifests are maintained in `exercises/3.7/`.

## Branch-Based Deployment Architecture

The deployment workflow is triggered whenever a commit is pushed to any branch.

Each branch receives an independent Kubernetes namespace.

| Git branch    | Kubernetes namespace |
| ------------- | -------------------- |
| `main`        | `project`            |
| `feature-a`   | `feature-a`          |
| `feature-b`   | `feature-b`          |
| `development` | `development`        |

The `main` branch continues to use the existing `project` namespace, preserving the deployment established in previous exercises.

Other branches use their branch name as the namespace.

Branch names must be valid Kubernetes namespace names.

## CI/CD Architecture

The deployment process is automated through GitHub Actions.

Whenever a commit is pushed to a branch, the workflow:

1. Checks out the repository.
2. Authenticates with Google Cloud using Workload Identity Federation.
3. Configures Docker authentication for Artifact Registry.
4. Retrieves credentials for the GKE cluster.
5. Determines the Kubernetes namespace from the branch name.
6. Creates the namespace if it does not already exist.
7. Builds the three application Docker images.
8. Pushes the images to Google Artifact Registry.
9. Updates the image references and namespace in the Kustomize configuration.
10. Applies the Kubernetes resources to GKE.
11. Verifies the frontend and backend Deployment rollouts and the Todo generator CronJob.

The workflow is defined in:

```text
.github/workflows/main.yaml
```

## Google Artifact Registry

The Docker repository is hosted in Google Artifact Registry.

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

Each image is tagged using the branch name and the full Git commit SHA.

For example:

```text
feature-a-COMMIT_SHA
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
- A repository restriction for `fjpiedade/DevOpsWithKubernetes`.
- A branch restriction allowing Git references beginning with `refs/heads/`.

The Workload Identity provider condition is:

```text
assertion.repository == 'fjpiedade/DevOpsWithKubernetes' && assertion.ref.startsWith('refs/heads/')
```

This permits authentication from any branch in the intended GitHub repository while excluding tags and other repositories.

The service account uses the following IAM roles:

- `roles/artifactregistry.writer` – allows publishing images to Artifact Registry.
- `roles/container.clusterViewer` – allows retrieving GKE cluster information and credentials.
- `roles/container.developer` – allows managing Kubernetes resources.

The deployment identity must also have sufficient Kubernetes permissions to create namespaces at cluster scope.

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

The workflow is triggered by pushes to all branches.

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

The workflow determines the namespace according to the branch name:

```bash
if [ "$BRANCH_NAME" = "main" ]; then
  NAMESPACE="project"
else
  NAMESPACE="$BRANCH_NAME"
fi
```

The namespace is then created if it does not already exist:

```bash
kubectl get namespace "$NAMESPACE" >/dev/null 2>&1 || \
  kubectl create namespace "$NAMESPACE"
```

This makes deployment repeatable because subsequent pushes to the same branch reuse its existing namespace.

The workflow builds and pushes the three Docker images and deploys the resulting versions using Kustomize.

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

The namespace is assigned dynamically during the workflow execution:

```bash
kustomize edit set namespace "$NAMESPACE"
```

Kustomize applies the selected namespace to the Kubernetes resources before deployment.

The namespace itself is created separately by the workflow and is not included as a resource in the Kustomize configuration.

### Project structure

```text
.github/
└── workflows/
    └── main.yaml

exercises/3.7/
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
├── todo-backend/
├── todo-generator/
└── README.md
```

### Render the Kustomize configuration

From the repository root, inspect the resources generated by Kustomize:

```bash
kubectl kustomize exercises/3.7/
```

This renders the Kubernetes configuration without applying it to the cluster.

The namespace can be changed dynamically for each deployment environment.

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

Each namespace has its own namespaced PVC and PostgreSQL resources.

## Automatic Deployment

### Prerequisites

- A Google Cloud project with billing enabled.
- A running Google Kubernetes Engine cluster.
- An Artifact Registry Docker repository.
- The Workload Identity Federation configuration allowing the intended GitHub repository branches.
- The GitHub repository variables configured.
- Sufficient Kubernetes permissions for the deployment identity, including namespace creation.
- GitHub Actions enabled for the repository.

### Trigger the deployment

From the repository root, commit and push changes to any branch.

For example, to deploy the `main` branch:

```bash
git add .github/workflows/main.yaml exercises/3.7/
git commit -m "Complete exercise 3.7: deploy separate environments per branch"
git push origin main
```

The push automatically triggers the GitHub Actions workflow.

A push to `main` deploys to the `project` namespace.

A push to another branch deploys to a namespace named after that branch.

### Monitor the workflow

Open the GitHub repository and navigate to:

**Actions → Build and deploy to GKE**

The workflow should complete all steps successfully.

### Verify the deployed resources

For the main environment:

```bash
kubectl get deployments -n project
kubectl get pods -n project
kubectl get cronjob -n project
kubectl get services -n project
kubectl get ingress -n project
```

For another branch, replace `feature-a` with the appropriate branch name:

```bash
kubectl get deployments -n feature-a
kubectl get pods -n feature-a
kubectl get cronjob -n feature-a
kubectl get services -n feature-a
kubectl get ingress -n feature-a
```

The frontend and backend Deployments should complete their rollouts successfully in each environment.

## Application Validation

### Verify the deployed images

Inspect the images used by the frontend and backend Deployments:

```bash
kubectl get deployments -n feature-a \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```

The application images should reference the Google Artifact Registry repository and the corresponding branch and commit tag.

### Verify the application endpoints

Retrieve the external Ingress IP for the selected environment:

```bash
kubectl get ingress todo-app-ingress -n feature-a
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

Each environment has its own Ingress. GKE may provision a separate external Load Balancer for each Ingress, which can incur additional costs.

## Result

Exercise 3.7 extends the Kustomize-based GKE deployment from Exercise 3.6 with branch-based environment isolation.

GitHub Actions builds and publishes the Todo frontend, backend and generator images to Google Artifact Registry.

Workload Identity Federation provides secure authentication between GitHub Actions and Google Cloud without requiring long-lived service account keys.

The `main` branch continues to use the `project` namespace, while other branches are deployed to independent namespaces matching their branch names.

Kustomize dynamically assigns the appropriate namespace to all application resources.

This allows different branches to be deployed and tested independently without replacing the application resources deployed by another branch.

## Cleanup

Google Cloud resources such as GKE clusters, persistent disks, Ingress resources and external Load Balancers may incur charges.

To remove an environment belonging to a non-main branch, delete its namespace:

```bash
kubectl delete namespace feature-a
```

Replace `feature-a` with the appropriate branch namespace.

Deleting a namespace removes its namespaced application resources, including Deployments, Services, Secrets, PVCs and Ingress resources.

The `project` namespace must not be deleted when cleaning up a feature branch. It is the main environment and is retained from previous exercises.

After completing the exercise and saving the required repository changes, remove any unused GKE resources if the cluster is no longer needed.

Verify that no unused Load Balancer resources or persistent disks remain in the Google Cloud project.
