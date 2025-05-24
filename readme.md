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

-   `PORT_MAPPINGS`: Additional port name to NodePort mappings in the format "name:port, name:port, ...". Example: "web:30080, web:30443".
-   `LABEL_SELECTOR`: Filter services by labels in the format "key1=value1,key2=value2". Example: "app=traefik,tier=frontend".

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

2.  **Publish:**
    ```bash
    make push
    ```
    You'll need to be logged into `ghcr.io` (`docker login ghcr.io`).

## Deploying with vcluster

To use this plugin with vcluster, you need to configure it in your vcluster Helm chart values or `vcluster.yaml`.

See the example configuration in `deploy/plugin-config-example.yaml`.

```bash
vcluster create my-vcluster -n my-namespace -f deploy/plugin-config-example.yaml
```

This command creates a new vcluster named "my-vcluster" in the "my-namespace" namespace, using the plugin configuration defined in the YAML file.