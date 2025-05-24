# Vcluster NodePort Patcher Plugin

This vcluster plugin connects to the host Kubernetes cluster and patches the `nodePort`
for specified ports ("web" and "websecure") of a target Service.

## Features

-   Reads target service details and desired NodePorts from environment variables.
-   Connects to the host Kubernetes cluster using the vcluster SDK.
-   Fetches the target Service by name and namespace.
-   Robustly finds ports named "web" and "websecure" to patch their `nodePort`.
-   Applies a JSON patch to update the `nodePort` values.
-   Logs its actions for traceability.
-   Performs the patch operation once on startup.

## Configuration

The plugin is configured via environment variables:

-   `TARGET_SERVICE_NAMESPACE`: (Required) The namespace of the Service to patch on the host cluster.
-   `TARGET_SERVICE_NAME`: (Required) The name of the Service to patch on the host cluster.
-   `HTTP_NODE_PORT`: (Required) The desired `nodePort` for the port named "web".
-   `HTTPS_NODE_PORT`: (Required) The desired `nodePort` for the port named "websecure".
-   `LOG_LEVEL`: (Optional) Set the log level. Defaults to "info". Can be "debug", "info", "warn", "error".

## Prerequisites

-   Go (version 1.20+ recommended)
-   Docker
-   Make (optional, for using the Makefile)
-   A running Kubernetes cluster where vcluster will be deployed.
-   `vcluster` CLI or Helm for deploying vcluster.

## Building the Plugin

### Using Makefile

1.  **Build the Go binary:**
    ```bash
    make build
    ```

2.  **Build the Docker image:**
    Replace `your-ghcr-username/your-repo-name` with your GitHub container registry username and repository name.
    ```bash
    IMAGE_NAME=ghcr.io/your-ghcr-username/your-repo-name/vcluster-nodeport-patcher make docker-build
    ```

3.  **Push the Docker image:**
    ```bash
    IMAGE_NAME=ghcr.io/your-ghcr-username/your-repo-name/vcluster-nodeport-patcher make docker-push
    ```
    You'll need to be logged into `ghcr.io` (`docker login ghcr.io`).

### Manual Build

1.  **Build the Go binary:**
    ```bash
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/plugin ./cmd/plugin/main.go
    ```

2.  **Build the Docker image:**
    ```bash
    docker build -t ghcr.io/your-ghcr-username/your-repo-name/vcluster-nodeport-patcher:latest .
    ```

## Deploying with vcluster

To use this plugin with vcluster, you need to configure it in your vcluster Helm chart values or `vcluster.yaml`.

See the example configuration in `deploy/plugin-config-example.yaml`. This example shows how to:
-   Specify the plugin image.
-   Set the required environment variables.
-   Define the necessary RBAC permissions for the plugin to interact with the host cluster.

**Key RBAC Permissions Required on Host Cluster:**

The vcluster service account (which the plugin will use) needs permissions to `get` and `patch` services in the target namespace on the host cluster.

Example `Role` and `RoleBinding` (or `ClusterRole` and `ClusterRoleBinding` if patching across namespaces not managed by vcluster's default host SA permissions):

```yaml
# This would be applied to the HOST cluster, granting permissions
# to the vcluster's service account in its host namespace.
# Adjust 'vcluster-sa' and 'vcluster-host-namespace' as needed.
# The plugin configuration in vcluster Helm values can also add these rules
# directly to the Role/ClusterRole managed by vcluster for its plugins.

apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  # Namespace where the target service exists, or the vcluster host namespace
  # if the plugin's default SA is used and it's the same.
  namespace: <TARGET_SERVICE_NAMESPACE_ON_HOST>
  name: vcluster-plugin-nodeport-patcher-role
rules:
- apiGroups: [""] # Core API group
  resources: ["services"]
  verbs: ["get", "patch"]
  # Optional: restrict to the specific service name for tighter security
  # resourceNames: ["<YOUR_TRAEFIK_SERVICE_NAME>"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: <TARGET_SERVICE_NAMESPACE_ON_HOST>
  name: vcluster-plugin-nodeport-patcher-binding
subjects:
- kind: ServiceAccount
  name: <VCLUSTER_SERVICE_ACCOUNT_NAME_IN_HOST_NAMESPACE> # e.g., vcluster, or the specific SA for your vcluster instance
  namespace: <VCLUSTER_HOST_NAMESPACE> # The namespace where vcluster itself is running on the host
roleRef:
  kind: Role
  name: vcluster-plugin-nodeport-patcher-role
  apiGroup: rbac.authorization.k8s.io