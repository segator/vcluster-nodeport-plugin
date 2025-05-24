package main

import (
	"context"
	"encoding/json"
	"fmt"
	"vcluster-nodeport-plugin/hook"
	"vcluster-nodeport-plugin/syncer"

	"os"
	"strconv"
	"strings"

	"github.com/loft-sh/vcluster-sdk/plugin"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// JSONPatchOperation describes a single JSON patch operation.
type JSONPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func main() {
	// Initialize plugin. ctx.PhysicalClusterConfig gives access to host cluster.
	ctx := plugin.MustInit() // This also initializes logging.

	// Get configuration from environment variables
	targetNamespace := os.Getenv("TARGET_SERVICE_NAMESPACE")
	targetService := os.Getenv("TARGET_SERVICE_NAME")
	httpNodePortStr := os.Getenv("HTTP_NODE_PORT")
	httpsNodePortStr := os.Getenv("HTTPS_NODE_PORT")

	// Validate required configuration
	if targetNamespace == "" {
		log.Fatal("Missing required environment variable: TARGET_SERVICE_NAMESPACE")
	}
	if targetService == "" {
		log.Fatal("Missing required environment variable: TARGET_SERVICE_NAME")
	}
	if httpNodePortStr == "" {
		log.Fatal("Missing required environment variable: HTTP_NODE_PORT")
	}
	if httpsNodePortStr == "" {
		log.Fatal("Missing required environment variable: HTTPS_NODE_PORT")
	}

	httpNodePort, err := strconv.Atoi(httpNodePortStr)
	if err != nil {
		log.Fatalf("Invalid HTTP_NODE_PORT value '%s': %v", httpNodePortStr, err)
	}
	httpsNodePort, err := strconv.Atoi(httpsNodePortStr)
	if err != nil {
		log.Fatalf("Invalid HTTPS_NODE_PORT value '%s': %v", httpsNodePortStr, err)
	}

	log.Infof("Plugin initialized. Target Service: %s/%s, HTTP NodePort: %d, HTTPS NodePort: %d",
		targetNamespace, targetService, httpNodePort, httpsNodePort)

	//plugin.MustRegister(syncer.NewServiceSyncer(ctx))
	plugin.MustRegister(hook.NewServiceHook())
	plugin.MustStart()
	// Create a Kubernetes clientset for the physical (host) cluster
	physicalClusterClient, err := kubernetes.NewForConfig(ctx.PhysicalClusterConfig)
	if err != nil {
		plugin.Log.Fatalf("Error creating physical cluster client: %v", err)
	}

	// Get the target service from the host cluster
	plugin.Log.Debugf("Fetching service %s/%s from host cluster...", targetNamespace, targetService)
	svc, err := physicalClusterClient.CoreV1().Services(targetNamespace).Get(context.TODO(), targetService, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			plugin.Log.Errorf("Service %s/%s not found in host cluster. Skipping patch.", targetNamespace, targetService)
			// Keep plugin running, maybe the service will be created later if this were a controller.
			// For a one-shot, this means it does nothing.
			plugin.MustStart() // Keep the plugin alive
			return
		}
		plugin.Log.Fatalf("Error getting service %s/%s from host cluster: %v", targetNamespace, targetService, err)
	}
	plugin.Log.Debugf("Successfully fetched service %s/%s.", targetNamespace, targetService)

	var patchOps []JSONPatchOperation
	webPortPatched := false
	websecurePortPatched := false

	for i, port := range svc.Spec.Ports {
		currentPortPath := fmt.Sprintf("/spec/ports/%d/nodePort", i)
		if strings.ToLower(port.Name) == "web" {
			if port.NodePort != int32(httpNodePort) {
				plugin.Log.Infof("Port 'web' (current NodePort: %d) will be patched to NodePort: %d", port.NodePort, httpNodePort)
				patchOps = append(patchOps, JSONPatchOperation{
					Op:    "replace",
					Path:  currentPortPath,
					Value: httpNodePort,
				})
			} else {
				plugin.Log.Infof("Port 'web' already has desired NodePort: %d. No patch needed for this port.", httpNodePort)
			}
			webPortPatched = true
		} else if strings.ToLower(port.Name) == "websecure" {
			if port.NodePort != int32(httpsNodePort) {
				plugin.Log.Infof("Port 'websecure' (current NodePort: %d) will be patched to NodePort: %d", port.NodePort, httpsNodePort)
				patchOps = append(patchOps, JSONPatchOperation{
					Op:    "replace",
					Path:  currentPortPath,
					Value: httpsNodePort,
				})
			} else {
				plugin.Log.Infof("Port 'websecure' already has desired NodePort: %d. No patch needed for this port.", httpsNodePort)
			}
			websecurePortPatched = true
		}
	}

	if !webPortPatched {
		plugin.Log.Warnf("Port named 'web' not found in service %s/%s. Cannot patch HTTP NodePort.", targetNamespace, targetService)
	}
	if !websecurePortPatched {
		plugin.Log.Warnf("Port named 'websecure' not found in service %s/%s. Cannot patch HTTPS NodePort.", targetNamespace, targetService)
	}

	if len(patchOps) == 0 {
		plugin.Log.Info("No changes required for service NodePorts. Exiting patch logic.")
		plugin.MustStart() // Keep the plugin alive
		return
	}

	patchBytes, err := json.Marshal(patchOps)
	if err != nil {
		plugin.Log.Fatalf("Error marshalling JSON patch operations: %v", err)
	}

	plugin.Log.Infof("Attempting to patch service %s/%s on host cluster with payload: %s", targetNamespace, targetService, string(patchBytes))
	_, err = physicalClusterClient.CoreV1().Services(targetNamespace).Patch(
		context.TODO(),
		targetService,
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		plugin.Log.Fatalf("Error patching service %s/%s on host cluster: %v", targetNamespace, targetService, err)
	}

	plugin.Log.Infof("Successfully patched service %s/%s on host cluster.", targetNamespace, targetService)

	// plugin.MustStart() will block and keep the plugin running.
	// For a one-shot task that runs on startup, this ensures the plugin
	// doesn't immediately exit and get restarted by Kubernetes if not desired.
	// It also allows the plugin to participate in leader election if it were a controller.
	plugin.Log.Info("Patch operation complete. Plugin will remain running.")
	plugin.MustStart()
}
