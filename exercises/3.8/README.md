# Exercise 3.8 – The Project, Step 17: Delete Branch Environments

## Objective

Automatically delete a Kubernetes environment when its corresponding Git branch is deleted.

## Implementation

The workflow `.github/workflows/delete-environment.yaml` is triggered by GitHub's `delete` event.

When a branch is deleted, the workflow:

1. Checks that the deleted reference is a branch and is not `main`.
2. Authenticates to Google Cloud using Workload Identity Federation.
3. Retrieves the credentials for the GKE cluster.
4. Deletes the Kubernetes namespace associated with the branch.

The namespace name matches the branch name. The `main` branch uses the permanent `project` namespace, which is protected from deletion.

The workflow uses `kubectl delete namespace` with `--ignore-not-found=true`, so it also handles cases where the namespace does not exist.

## Verification

The workflow was tested using a temporary branch named `feature-delete-test`.

1. Created and pushed the branch.
2. Verified that the deployment workflow created the `feature-delete-test` namespace.
3. Confirmed that the namespace was active in the GKE cluster.
4. Deleted the remote branch.
5. Verified that the delete workflow removed the namespace.
6. Confirmed that the namespace no longer existed.

The test branch was also deleted locally.

## Result

Branch environments are now automatically removed when their branches are deleted. The `main` branch and its `project` namespace remain unaffected.
