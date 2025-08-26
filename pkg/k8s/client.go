package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	sigsyaml "sigs.k8s.io/yaml"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client with auto-detection for various cluster types
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error
	var configSource string

	fmt.Println("🔍 Auto-detecting Kubernetes cluster configuration...")

	// Priority order for configuration detection:
	// 1. In-cluster config (highest priority for pod deployment)
	// 2. Environment variables
	// 3. K3s default location
	// 4. Standard kubeconfig locations
	// 5. Development fallbacks

	// Method 1: In-cluster configuration (for pods running in cluster)
	if isRunningInCluster() {
		fmt.Println("📦 Detected running inside Kubernetes cluster")
		config, err = rest.InClusterConfig()
		if err == nil {
			configSource = "in-cluster"
			fmt.Println("✅ Successfully loaded in-cluster configuration")
		} else {
			fmt.Printf("⚠️  In-cluster config failed: %v\n", err)
		}
	}

	// Method 2: KUBECONFIG environment variable
	if config == nil {
		if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
			fmt.Printf("🔧 Found KUBECONFIG environment variable: %s\n", kubeconfigPath)
			if _, err := os.Stat(kubeconfigPath); err == nil {
				config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
				if err == nil {
					configSource = "KUBECONFIG env var"
					fmt.Printf("✅ Successfully loaded config from KUBECONFIG: %s\n", kubeconfigPath)
				} else {
					fmt.Printf("⚠️  Failed to load KUBECONFIG: %v\n", err)
				}
			} else {
				fmt.Printf("⚠️  KUBECONFIG file not found: %s\n", kubeconfigPath)
			}
		}
	}

	// Method 3: K3s default locations (multiple possible paths)
	if config == nil {
		k3sPaths := []string{
			"/etc/rancher/k3s/k3s.yaml",
			"/var/lib/rancher/k3s/server/cred/admin.kubeconfig",
			"/etc/kubernetes/admin.conf", // Some K3s installations
		}

		for _, k3sPath := range k3sPaths {
			if _, err := os.Stat(k3sPath); err == nil {
				fmt.Printf("🐄 Found Kubernetes kubeconfig at: %s\n", k3sPath)
				config, err = clientcmd.BuildConfigFromFlags("", k3sPath)
				if err == nil {
					configSource = fmt.Sprintf("Kubernetes config (%s)", k3sPath)
					fmt.Printf("✅ Successfully loaded Kubernetes configuration\n")
					break
				} else {
					fmt.Printf("⚠️  Failed to load Kubernetes config from %s: %v\n", k3sPath, err)
				}
			}
		}
	}

	// Method 4: Standard kubeconfig locations
	if config == nil {
		standardPaths := []string{}

		if home := homedir.HomeDir(); home != "" {
			standardPaths = append(standardPaths,
				filepath.Join(home, ".kube", "config"),
				filepath.Join(home, ".kube", "config.yaml"),
			)
		}

		// Add system-wide locations
		standardPaths = append(standardPaths,
			"/root/.kube/config",
			"/home/kubernetes/.kube/config",
		)

		for _, stdPath := range standardPaths {
			if _, err := os.Stat(stdPath); err == nil {
				fmt.Printf("📁 Found standard kubeconfig at: %s\n", stdPath)
				config, err = clientcmd.BuildConfigFromFlags("", stdPath)
				if err == nil {
					configSource = fmt.Sprintf("Standard config (%s)", stdPath)
					fmt.Printf("✅ Successfully loaded standard configuration\n")
					break
				} else {
					fmt.Printf("⚠️  Failed to load standard config from %s: %v\n", stdPath, err)
				}
			}
		}
	}

	// Method 5: Try to auto-create from service account (K8s cluster)
	if config == nil {
		fmt.Println("🔄 Attempting to create config from service account...")
		config, err = createConfigFromServiceAccount()
		if err == nil {
			configSource = "service account auto-config"
			fmt.Println("✅ Successfully created config from service account")
		} else {
			fmt.Printf("⚠️  Service account config failed: %v\n", err)
		}
	}

	// If all methods failed, return error with helpful information
	if config == nil {
		return nil, fmt.Errorf(`
❌ Failed to find Kubernetes configuration in any location.

Tried the following locations:
  1. In-cluster config (for pods)
  2. KUBECONFIG environment variable
  3. K3s locations: /etc/rancher/k3s/k3s.yaml
  4. Standard locations: ~/.kube/config
  5. Service account auto-configuration

To fix this issue:
  • For K3s: Set KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  • For K8s: Ensure ~/.kube/config exists
  • For containers: Mount kubeconfig or use service account
  • Set environment: K8S_AUTO_CONFIG=true for development

Error details: %v`, err)
	}

	// Enhanced configuration for different cluster types
	enhanceConfigForClusterType(config, configSource)

	// Test the configuration
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset with %s: %v", configSource, err)
	}

	// Final connectivity test
	client := &Client{clientset: clientset}
	if err := client.TestConnection(); err != nil {
		// If connection fails, try with relaxed TLS settings for development
		if isDevelopmentMode() {
			fmt.Println("🔧 Connection failed, trying with relaxed TLS settings for development...")
			config.TLSClientConfig.Insecure = true
			clientset, err = kubernetes.NewForConfig(config)
			if err == nil {
				client = &Client{clientset: clientset}
				if err := client.TestConnection(); err == nil {
					fmt.Println("⚠️  Connected with insecure TLS (development mode only)")
					configSource += " (insecure)"
				} else {
					return nil, fmt.Errorf("connection failed even with relaxed TLS settings: %v", err)
				}
			}
		} else {
			return nil, fmt.Errorf("failed to connect to Kubernetes cluster using %s: %v", configSource, err)
		}
	}

	fmt.Printf("🎉 Successfully connected to Kubernetes cluster using: %s\n", configSource)
	return client, nil
}

// isRunningInCluster detects if we're running inside a Kubernetes pod
func isRunningInCluster() bool {
	// Check for service account token
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	// Check for Kubernetes environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		return true
	}

	return false
}

// createConfigFromServiceAccount attempts to create config from mounted service account
func createConfigFromServiceAccount() (*rest.Config, error) {
	tokenFile := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	caFile := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	if _, err := os.Stat(tokenFile); err != nil {
		return nil, fmt.Errorf("service account token not found")
	}

	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")

	if host == "" || port == "" {
		return nil, fmt.Errorf("kubernetes service environment variables not found")
	}

	config := &rest.Config{
		Host: fmt.Sprintf("https://%s:%s", host, port),
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: caFile,
		},
		BearerTokenFile: tokenFile,
	}

	return config, nil
}

// enhanceConfigForClusterType applies cluster-specific optimizations
func enhanceConfigForClusterType(config *rest.Config, configSource string) {
	// Set reasonable timeouts
	config.Timeout = 30 * time.Second

	// Apply cluster-specific settings
	if strings.Contains(strings.ToLower(configSource), "k3s") {
		fmt.Println("🐄 Applying K3s-specific optimizations...")
		// K3s often has longer certificate chains
		config.TLSClientConfig.ServerName = ""

		// For development environments, allow some flexibility
		if isDevelopmentMode() {
			fmt.Println("🔧 Development mode: Relaxing TLS settings for K3s")
			config.TLSClientConfig.Insecure = false // Keep secure but flexible
		}
	} else if strings.Contains(strings.ToLower(configSource), "in-cluster") {
		fmt.Println("📦 Applying in-cluster optimizations...")
		// In-cluster connections are typically more reliable
		config.QPS = 100
		config.Burst = 200
	} else {
		fmt.Println("☸️  Applying standard Kubernetes optimizations...")
		// Standard K8s cluster settings
		config.QPS = 50
		config.Burst = 100
	}
}

// isDevelopmentMode checks if we're in development mode
func isDevelopmentMode() bool {
	return os.Getenv("K8S_AUTO_CONFIG") == "true" ||
		os.Getenv("DEVELOPMENT_MODE") == "true" ||
		os.Getenv("K3S_INSECURE_SKIP_VERIFY") == "true"
}

// Enhanced TestConnection with better error reporting
func (c *Client) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: Get server version
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to get server version: %v", err)
	}
	fmt.Printf("📋 Connected to Kubernetes %s\n", version.String())

	// Test 2: Try to list namespaces (basic permission test)
	_, err = c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to list namespaces (permission test): %v", err)
	}

	fmt.Println("✅ Basic connectivity and permissions verified")
	return nil
}

// ========== NAMESPACE OPERATIONS ==========

// ListNamespaces returns a list of all namespaces in the cluster
func (c *Client) ListNamespaces(ctx context.Context) ([]map[string]interface{}, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var result []map[string]interface{}
	for _, ns := range namespaces.Items {
		nsInfo := map[string]interface{}{
			"name":              ns.Name,
			"status":            string(ns.Status.Phase),
			"creationTimestamp": ns.CreationTimestamp.Time,
			"labels":            ns.Labels,
			"annotations":       ns.Annotations,
		}
		result = append(result, nsInfo)
	}

	return result, nil
}

// GetNamespace returns detailed information about a specific namespace
func (c *Client) GetNamespace(ctx context.Context, name string) (map[string]interface{}, error) {
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"name":              namespace.Name,
		"status":            string(namespace.Status.Phase),
		"creationTimestamp": namespace.CreationTimestamp.Time,
		"labels":            namespace.Labels,
		"annotations":       namespace.Annotations,
		"resourceVersion":   namespace.ResourceVersion,
		"uid":               string(namespace.UID),
	}

	return result, nil
}

// CreateNamespace creates a new namespace with optional labels and annotations
func (c *Client) CreateNamespace(ctx context.Context, name string, labels, annotations map[string]string) (map[string]interface{}, error) {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}

	createdNs, err := c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"name":              createdNs.Name,
		"status":            string(createdNs.Status.Phase),
		"creationTimestamp": createdNs.CreationTimestamp.Time,
		"labels":            createdNs.Labels,
		"annotations":       createdNs.Annotations,
		"resourceVersion":   createdNs.ResourceVersion,
		"uid":               string(createdNs.UID),
	}

	return result, nil
}

// UpdateNamespace updates labels and annotations of an existing namespace
func (c *Client) UpdateNamespace(ctx context.Context, name string, labels, annotations map[string]string) (map[string]interface{}, error) {
	// Get the current namespace
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace '%s': %v", name, err)
	}

	// Update labels and annotations
	if labels != nil {
		namespace.Labels = labels
	}
	if annotations != nil {
		namespace.Annotations = annotations
	}

	// Apply the update
	updatedNs, err := c.clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update namespace '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"name":              updatedNs.Name,
		"status":            string(updatedNs.Status.Phase),
		"creationTimestamp": updatedNs.CreationTimestamp.Time,
		"labels":            updatedNs.Labels,
		"annotations":       updatedNs.Annotations,
		"resourceVersion":   updatedNs.ResourceVersion,
		"uid":               string(updatedNs.UID),
	}

	return result, nil
}

// DeleteNamespace deletes a namespace (this will also delete all resources in it)
func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %v", name, err)
	}
	return nil
}

// GetNamespaceResourceQuota returns resource quotas for a namespace
func (c *Client) GetNamespaceResourceQuota(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	quotas, err := c.clientset.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quotas for namespace '%s': %v", namespace, err)
	}

	var result []map[string]interface{}
	for _, quota := range quotas.Items {
		quotaInfo := map[string]interface{}{
			"name":              quota.Name,
			"namespace":         quota.Namespace,
			"hard":              quota.Status.Hard,
			"used":              quota.Status.Used,
			"creationTimestamp": quota.CreationTimestamp.Time,
		}
		result = append(result, quotaInfo)
	}

	return result, nil
}

// Add these new methods to your client.go file

// GetNamespaceEvents returns events related to a specific namespace
func (c *Client) GetNamespaceEvents(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for namespace '%s': %v", namespace, err)
	}

	var result []map[string]interface{}
	for _, event := range events.Items {
		eventInfo := map[string]interface{}{
			"type":           event.Type,
			"reason":         event.Reason,
			"message":        event.Message,
			"firstTimestamp": event.FirstTimestamp.Time,
			"lastTimestamp":  event.LastTimestamp.Time,
			"count":          event.Count,
			"source":         event.Source.Component,
			"object":         fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
		}
		result = append(result, eventInfo)
	}

	return result, nil
}

// GetNamespaceAllResources returns all resources in a namespace to help identify what's blocking deletion
func (c *Client) GetNamespaceAllResources(ctx context.Context, namespace string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"namespace": namespace,
		"resources": map[string]interface{}{},
	}

	// Get pods
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(pods.Items) > 0 {
		var podList []map[string]interface{}
		for _, pod := range pods.Items {
			podInfo := map[string]interface{}{
				"name":       pod.Name,
				"status":     string(pod.Status.Phase),
				"finalizers": pod.Finalizers,
			}
			podList = append(podList, podInfo)
		}
		result["resources"].(map[string]interface{})["pods"] = podList
	}

	// Get services
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(services.Items) > 0 {
		var serviceList []map[string]interface{}
		for _, svc := range services.Items {
			serviceInfo := map[string]interface{}{
				"name":       svc.Name,
				"type":       string(svc.Spec.Type),
				"finalizers": svc.Finalizers,
			}
			serviceList = append(serviceList, serviceInfo)
		}
		result["resources"].(map[string]interface{})["services"] = serviceList
	}

	// Get deployments
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(deployments.Items) > 0 {
		var deploymentList []map[string]interface{}
		for _, deploy := range deployments.Items {
			deployInfo := map[string]interface{}{
				"name":       deploy.Name,
				"replicas":   *deploy.Spec.Replicas,
				"ready":      deploy.Status.ReadyReplicas,
				"finalizers": deploy.Finalizers,
			}
			deploymentList = append(deploymentList, deployInfo)
		}
		result["resources"].(map[string]interface{})["deployments"] = deploymentList
	}

	// Get persistent volume claims
	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(pvcs.Items) > 0 {
		var pvcList []map[string]interface{}
		for _, pvc := range pvcs.Items {
			pvcInfo := map[string]interface{}{
				"name":       pvc.Name,
				"status":     string(pvc.Status.Phase),
				"finalizers": pvc.Finalizers,
			}
			pvcList = append(pvcList, pvcInfo)
		}
		result["resources"].(map[string]interface{})["persistentVolumeClaims"] = pvcList
	}

	// Get secrets
	secrets, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(secrets.Items) > 0 {
		var secretList []map[string]interface{}
		for _, secret := range secrets.Items {
			secretInfo := map[string]interface{}{
				"name":       secret.Name,
				"type":       string(secret.Type),
				"finalizers": secret.Finalizers,
			}
			secretList = append(secretList, secretInfo)
		}
		result["resources"].(map[string]interface{})["secrets"] = secretList
	}

	return result, nil
}

// ForceDeleteNamespace attempts to force delete a namespace using multiple strategies
func (c *Client) ForceDeleteNamespace(ctx context.Context, name string) error {
	// Strategy 1: Try regular delete first
	fmt.Printf("Attempting regular delete for namespace '%s'...\n", name)
	err := c.DeleteNamespace(ctx, name)
	if err == nil {
		// Wait and check if it's actually deleted
		if c.waitForNamespaceDeletion(ctx, name, 10*time.Second) {
			return nil
		}
	}

	// Strategy 2: Enhanced force delete with multiple approaches
	return c.enhancedForceDelete(ctx, name)
}

// enhancedForceDelete implements multiple strategies for stuck namespaces
func (c *Client) enhancedForceDelete(ctx context.Context, name string) error {
	fmt.Printf("Namespace '%s' requires force deletion...\n", name)

	// Get current namespace state
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to get namespace '%s': %v", name, err)
	}

	// Check current conditions
	fmt.Printf("Namespace status: %s\n", namespace.Status.Phase)
	if len(namespace.Status.Conditions) > 0 {
		fmt.Println("Namespace conditions:")
		for _, condition := range namespace.Status.Conditions {
			fmt.Printf("  - %s: %s (%s)\n", condition.Type, condition.Status, condition.Reason)
		}
	}

	// Strategy 2a: Remove spec finalizers
	if len(namespace.Spec.Finalizers) > 0 {
		fmt.Printf("Removing spec finalizers: %v\n", namespace.Spec.Finalizers)
		namespace.Spec.Finalizers = []corev1.FinalizerName{}

		_, err = c.clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Warning: Failed to remove spec finalizers: %v\n", err)
		} else {
			if c.waitForNamespaceDeletion(ctx, name, 15*time.Second) {
				return nil
			}
		}
	}

	// Strategy 2b: Remove metadata finalizers
	if len(namespace.ObjectMeta.Finalizers) > 0 {
		fmt.Printf("Removing metadata finalizers: %v\n", namespace.ObjectMeta.Finalizers)

		// Get fresh namespace state
		namespace, err = c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil
			}
			return fmt.Errorf("failed to get fresh namespace state: %v", err)
		}

		namespace.ObjectMeta.Finalizers = []string{}
		_, err = c.clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Warning: Failed to remove metadata finalizers: %v\n", err)
		} else {
			if c.waitForNamespaceDeletion(ctx, name, 15*time.Second) {
				return nil
			}
		}
	}

	// Strategy 2c: Use finalize subresource (K3s specific)
	fmt.Printf("Attempting finalize subresource approach...\n")
	err = c.finalizeNamespace(ctx, name)
	if err != nil {
		fmt.Printf("Warning: Finalize subresource failed: %v\n", err)
	} else {
		if c.waitForNamespaceDeletion(ctx, name, 10*time.Second) {
			return nil
		}
	}

	// Strategy 2d: Direct JSON patch (last resort)
	fmt.Printf("Attempting direct JSON patch...\n")
	err = c.patchNamespaceFinalizers(ctx, name)
	if err != nil {
		fmt.Printf("Warning: JSON patch failed: %v\n", err)
	} else {
		if c.waitForNamespaceDeletion(ctx, name, 10*time.Second) {
			return nil
		}
	}

	// Final check
	_, err = c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil // Successfully deleted
	}

	return fmt.Errorf("namespace '%s' could not be force deleted after trying all strategies", name)
}

// waitForNamespaceDeletion waits for a namespace to be deleted
func (c *Client) waitForNamespaceDeletion(ctx context.Context, name string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil && strings.Contains(err.Error(), "not found") {
			fmt.Printf("Namespace '%s' successfully deleted\n", name)
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// finalizeNamespace uses the finalize subresource
func (c *Client) finalizeNamespace(ctx context.Context, name string) error {
	// Get current namespace
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Clear finalizers and update status
	namespace.Spec.Finalizers = []corev1.FinalizerName{}
	namespace.ObjectMeta.Finalizers = []string{}

	// Update the finalize subresource
	_, err = c.clientset.CoreV1().Namespaces().UpdateStatus(ctx, namespace, metav1.UpdateOptions{})
	return err
}

// patchNamespaceFinalizers uses JSON patch to remove finalizers
func (c *Client) patchNamespaceFinalizers(ctx context.Context, name string) error {
	// Create JSON patch to remove finalizers
	patch := []byte(`[
        {"op": "replace", "path": "/spec/finalizers", "value": []},
        {"op": "replace", "path": "/metadata/finalizers", "value": []}
    ]`)

	_, err := c.clientset.CoreV1().Namespaces().Patch(ctx, name, "application/json-patch+json", patch, metav1.PatchOptions{})
	return err
}

// GetNamespaceYAML returns the YAML definition of a namespace
func (c *Client) GetNamespaceYAML(ctx context.Context, name string) (string, error) {
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get namespace '%s': %v", name, err)
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(namespace)
	if err != nil {
		return "", fmt.Errorf("failed to convert namespace to YAML: %v", err)
	}

	return string(yamlData), nil
}

// SetNamespaceResourceQuota creates or updates a resource quota in a namespace
func (c *Client) SetNamespaceResourceQuota(ctx context.Context, namespace, manifest string) (map[string]interface{}, error) {
	// Parse the JSON manifest
	var resourceQuota corev1.ResourceQuota
	err := json.Unmarshal([]byte(manifest), &resourceQuota)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource quota manifest: %v", err)
	}

	// Ensure the namespace is set correctly
	if resourceQuota.Namespace == "" {
		resourceQuota.Namespace = namespace
	}
	if resourceQuota.Namespace != namespace {
		return nil, fmt.Errorf("resource quota namespace '%s' does not match target namespace '%s'", resourceQuota.Namespace, namespace)
	}

	// Try to get existing resource quota first
	existingQuota, err := c.clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, resourceQuota.Name, metav1.GetOptions{})
	if err == nil {
		// Update existing resource quota
		resourceQuota.ResourceVersion = existingQuota.ResourceVersion
		updatedQuota, err := c.clientset.CoreV1().ResourceQuotas(namespace).Update(ctx, &resourceQuota, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update resource quota: %v", err)
		}

		result := map[string]interface{}{
			"name":              updatedQuota.Name,
			"namespace":         updatedQuota.Namespace,
			"hard":              updatedQuota.Status.Hard,
			"used":              updatedQuota.Status.Used,
			"creationTimestamp": updatedQuota.CreationTimestamp.Time,
			"operation":         "updated",
		}
		return result, nil
	} else {
		// Create new resource quota
		createdQuota, err := c.clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, &resourceQuota, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create resource quota: %v", err)
		}

		result := map[string]interface{}{
			"name":              createdQuota.Name,
			"namespace":         createdQuota.Namespace,
			"hard":              createdQuota.Status.Hard,
			"used":              createdQuota.Status.Used,
			"creationTimestamp": createdQuota.CreationTimestamp.Time,
			"operation":         "created",
		}
		return result, nil
	}
}

// GetNamespaceLimitRanges returns limit ranges for a namespace
func (c *Client) GetNamespaceLimitRanges(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	limitRanges, err := c.clientset.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get limit ranges for namespace '%s': %v", namespace, err)
	}

	var result []map[string]interface{}
	for _, lr := range limitRanges.Items {
		limitRangeInfo := map[string]interface{}{
			"name":              lr.Name,
			"namespace":         lr.Namespace,
			"limits":            lr.Spec.Limits,
			"creationTimestamp": lr.CreationTimestamp.Time,
		}
		result = append(result, limitRangeInfo)
	}

	return result, nil
}

// SetNamespaceLimitRange creates or updates a limit range in a namespace
func (c *Client) SetNamespaceLimitRange(ctx context.Context, namespace, manifest string) (map[string]interface{}, error) {
	// Parse the JSON manifest
	var limitRange corev1.LimitRange
	err := json.Unmarshal([]byte(manifest), &limitRange)
	if err != nil {
		return nil, fmt.Errorf("failed to parse limit range manifest: %v", err)
	}

	// Ensure the namespace is set correctly
	if limitRange.Namespace == "" {
		limitRange.Namespace = namespace
	}
	if limitRange.Namespace != namespace {
		return nil, fmt.Errorf("limit range namespace '%s' does not match target namespace '%s'", limitRange.Namespace, namespace)
	}

	// Try to get existing limit range first
	existingLimitRange, err := c.clientset.CoreV1().LimitRanges(namespace).Get(ctx, limitRange.Name, metav1.GetOptions{})
	if err == nil {
		// Update existing limit range
		limitRange.ResourceVersion = existingLimitRange.ResourceVersion
		updatedLimitRange, err := c.clientset.CoreV1().LimitRanges(namespace).Update(ctx, &limitRange, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to update limit range: %v", err)
		}

		result := map[string]interface{}{
			"name":              updatedLimitRange.Name,
			"namespace":         updatedLimitRange.Namespace,
			"limits":            updatedLimitRange.Spec.Limits,
			"creationTimestamp": updatedLimitRange.CreationTimestamp.Time,
			"operation":         "updated",
		}
		return result, nil
	} else {
		// Create new limit range
		createdLimitRange, err := c.clientset.CoreV1().LimitRanges(namespace).Create(ctx, &limitRange, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create limit range: %v", err)
		}

		result := map[string]interface{}{
			"name":              createdLimitRange.Name,
			"namespace":         createdLimitRange.Namespace,
			"limits":            createdLimitRange.Spec.Limits,
			"creationTimestamp": createdLimitRange.CreationTimestamp.Time,
			"operation":         "created",
		}
		return result, nil
	}
}

// ========== POD OPERATIONS ==========
// GetPodsInNamespace returns detailed pod information in the specified namespace
func (c *Client) GetPodsInNamespace(namespace string) ([]map[string]interface{}, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	var result []map[string]interface{}
	for _, pod := range pods.Items {
		podInfo := map[string]interface{}{
			"name":              pod.Name,
			"namespace":         pod.Namespace,
			"status":            string(pod.Status.Phase),
			"nodeName":          pod.Spec.NodeName,
			"creationTimestamp": pod.CreationTimestamp.Time,
			"labels":            pod.Labels,
			"annotations":       pod.Annotations,
			"restartCount":      getPodRestartCount(&pod),
			"ready":             isPodReady(&pod),
			"containers":        getContainerInfo(&pod),
		}
		result = append(result, podInfo)
	}

	return result, nil
}

// GetPodsInNamespaceWithSelector returns pods filtered by label selector
func (c *Client) GetPodsInNamespaceWithSelector(namespace, labelSelector string) ([]map[string]interface{}, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	var result []map[string]interface{}
	for _, pod := range pods.Items {
		podInfo := map[string]interface{}{
			"name":              pod.Name,
			"namespace":         pod.Namespace,
			"status":            string(pod.Status.Phase),
			"nodeName":          pod.Spec.NodeName,
			"creationTimestamp": pod.CreationTimestamp.Time,
			"labels":            pod.Labels,
			"annotations":       pod.Annotations,
			"restartCount":      getPodRestartCount(&pod),
			"ready":             isPodReady(&pod),
			"containers":        getContainerInfo(&pod),
		}
		result = append(result, podInfo)
	}

	return result, nil
}

// GetPod returns detailed information about a specific pod
func (c *Client) GetPod(ctx context.Context, namespace, name string) (map[string]interface{}, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod '%s' in namespace '%s': %v", name, namespace, err)
	}

	result := map[string]interface{}{
		"name":              pod.Name,
		"namespace":         pod.Namespace,
		"status":            string(pod.Status.Phase),
		"statusMessage":     pod.Status.Message,
		"nodeName":          pod.Spec.NodeName,
		"hostIP":            pod.Status.HostIP,
		"podIP":             pod.Status.PodIP,
		"creationTimestamp": pod.CreationTimestamp.Time,
		"labels":            pod.Labels,
		"annotations":       pod.Annotations,
		"restartCount":      getPodRestartCount(pod),
		"ready":             isPodReady(pod),
		"containers":        getContainerInfo(pod),
		"conditions":        getPodConditions(pod),
		"volumes":           getVolumeInfo(pod),
		"resourceVersion":   pod.ResourceVersion,
		"uid":               string(pod.UID),
	}

	return result, nil
}

// GetPodLogs retrieves logs from a specific pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, name, containerName string, tailLines int64, follow, previous bool) (string, error) {
	logOptions := &corev1.PodLogOptions{
		Follow:     follow,
		Previous:   previous,
		TailLines:  &tailLines,
		Timestamps: true,
	}

	if containerName != "" {
		logOptions.Container = containerName
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, logOptions)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod '%s' in namespace '%s': %v", name, namespace, err)
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, logs); err != nil {
		return "", fmt.Errorf("failed to read logs: %v", err)
	}

	return buf.String(), nil
}

// DeletePod deletes a specific pod
func (c *Client) DeletePod(ctx context.Context, namespace, name string, gracePeriodSeconds int64) error {
	deleteOptions := metav1.DeleteOptions{}
	if gracePeriodSeconds > 0 {
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	err := c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete pod '%s' in namespace '%s': %v", name, namespace, err)
	}

	return nil
}

// GetPodEvents retrieves events related to a specific pod
func (c *Client) GetPodEvents(ctx context.Context, namespace, podName string) ([]map[string]interface{}, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for pod '%s': %v", podName, err)
	}

	var result []map[string]interface{}
	for _, event := range events.Items {
		eventInfo := map[string]interface{}{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"timestamp": event.FirstTimestamp.Time,
			"count":     event.Count,
			"source":    event.Source.Component,
		}
		result = append(result, eventInfo)
	}

	return result, nil
}

// Helper functions
func getPodRestartCount(pod *corev1.Pod) int32 {
	var totalRestarts int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += containerStatus.RestartCount
	}
	return totalRestarts
}

func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func getContainerInfo(pod *corev1.Pod) []map[string]interface{} {
	var containers []map[string]interface{}

	for _, container := range pod.Spec.Containers {
		containerInfo := map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
		}

		// Add status information if available
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == container.Name {
				containerInfo["ready"] = status.Ready
				containerInfo["restartCount"] = status.RestartCount
				if status.State.Running != nil {
					containerInfo["state"] = "running"
					containerInfo["startedAt"] = status.State.Running.StartedAt.Time
				} else if status.State.Waiting != nil {
					containerInfo["state"] = "waiting"
					containerInfo["reason"] = status.State.Waiting.Reason
				} else if status.State.Terminated != nil {
					containerInfo["state"] = "terminated"
					containerInfo["reason"] = status.State.Terminated.Reason
				}
				break
			}
		}

		containers = append(containers, containerInfo)
	}

	return containers
}

func getPodConditions(pod *corev1.Pod) []map[string]interface{} {
	var conditions []map[string]interface{}
	for _, condition := range pod.Status.Conditions {
		conditionInfo := map[string]interface{}{
			"type":               string(condition.Type),
			"status":             string(condition.Status),
			"reason":             condition.Reason,
			"message":            condition.Message,
			"lastTransitionTime": condition.LastTransitionTime.Time,
		}
		conditions = append(conditions, conditionInfo)
	}
	return conditions
}

func getVolumeInfo(pod *corev1.Pod) []map[string]interface{} {
	var volumes []map[string]interface{}
	for _, volume := range pod.Spec.Volumes {
		volumeInfo := map[string]interface{}{
			"name": volume.Name,
		}

		if volume.ConfigMap != nil {
			volumeInfo["type"] = "configMap"
			volumeInfo["configMapName"] = volume.ConfigMap.Name
		} else if volume.Secret != nil {
			volumeInfo["type"] = "secret"
			volumeInfo["secretName"] = volume.Secret.SecretName
		} else if volume.PersistentVolumeClaim != nil {
			volumeInfo["type"] = "persistentVolumeClaim"
			volumeInfo["claimName"] = volume.PersistentVolumeClaim.ClaimName
		} else if volume.EmptyDir != nil {
			volumeInfo["type"] = "emptyDir"
		}

		volumes = append(volumes, volumeInfo)
	}
	return volumes
}

// CreatePod creates a new pod from a JSON manifest
func (c *Client) CreatePod(ctx context.Context, namespace string, podManifest string) (map[string]interface{}, error) {
	// Parse the JSON manifest
	var pod corev1.Pod
	err := json.Unmarshal([]byte(podManifest), &pod)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pod manifest: %v", err)
	}

	// Ensure the namespace is set correctly
	if pod.Namespace == "" {
		pod.Namespace = namespace
	}
	if pod.Namespace != namespace {
		return nil, fmt.Errorf("pod namespace '%s' does not match target namespace '%s'", pod.Namespace, namespace)
	}

	// Create the pod
	createdPod, err := c.clientset.CoreV1().Pods(namespace).Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// Return the created pod information
	result := map[string]interface{}{
		"name":              createdPod.Name,
		"namespace":         createdPod.Namespace,
		"status":            string(createdPod.Status.Phase),
		"nodeName":          createdPod.Spec.NodeName,
		"creationTimestamp": createdPod.CreationTimestamp.Time,
		"labels":            createdPod.Labels,
		"annotations":       createdPod.Annotations,
		"containers":        getContainerInfo(createdPod),
		"resourceVersion":   createdPod.ResourceVersion,
		"uid":               string(createdPod.UID),
	}

	return result, nil
}

// UpdatePod updates an existing pod (limited to labels and annotations)
func (c *Client) UpdatePod(ctx context.Context, namespace, name string, labels, annotations map[string]string) (map[string]interface{}, error) {
	// Get the current pod
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod '%s' in namespace '%s': %v", name, namespace, err)
	}

	// Update labels if provided
	if labels != nil {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		for k, v := range labels {
			pod.Labels[k] = v
		}
	}

	// Update annotations if provided
	if annotations != nil {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		for k, v := range annotations {
			pod.Annotations[k] = v
		}
	}

	// Apply the update
	updatedPod, err := c.clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update pod '%s' in namespace '%s': %v", name, namespace, err)
	}

	// Return the updated pod information
	result := map[string]interface{}{
		"name":              updatedPod.Name,
		"namespace":         updatedPod.Namespace,
		"status":            string(updatedPod.Status.Phase),
		"nodeName":          updatedPod.Spec.NodeName,
		"creationTimestamp": updatedPod.CreationTimestamp.Time,
		"labels":            updatedPod.Labels,
		"annotations":       updatedPod.Annotations,
		"containers":        getContainerInfo(updatedPod),
		"resourceVersion":   updatedPod.ResourceVersion,
		"uid":               string(updatedPod.UID),
	}

	return result, nil
}

// ========== DEPLOYMENT OPERATIONS ==========

// ListDeployments returns a list of deployments in the specified namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace '%s': %v", namespace, err)
	}

	var result []map[string]interface{}
	for _, deployment := range deployments.Items {
		deploymentInfo := map[string]interface{}{
			"name":              deployment.Name,
			"namespace":         deployment.Namespace,
			"replicas":          *deployment.Spec.Replicas,
			"readyReplicas":     deployment.Status.ReadyReplicas,
			"availableReplicas": deployment.Status.AvailableReplicas,
			"updatedReplicas":   deployment.Status.UpdatedReplicas,
			"creationTimestamp": deployment.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":            deployment.Labels,
			"annotations":       deployment.Annotations,
			"selector":          deployment.Spec.Selector.MatchLabels,
			"strategy":          deployment.Spec.Strategy.Type,
			"conditions":        deployment.Status.Conditions,
		}

		// Add container information
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			var containers []map[string]interface{}
			for _, container := range deployment.Spec.Template.Spec.Containers {
				containerInfo := map[string]interface{}{
					"name":  container.Name,
					"image": container.Image,
				}
				if len(container.Ports) > 0 {
					containerInfo["ports"] = container.Ports
				}
				containers = append(containers, containerInfo)
			}
			deploymentInfo["containers"] = containers
		}

		result = append(result, deploymentInfo)
	}

	return result, nil
}

// ListDeploymentsWithSelector returns deployments filtered by label selector
func (c *Client) ListDeploymentsWithSelector(ctx context.Context, namespace, labelSelector string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments with selector '%s' in namespace '%s': %v", labelSelector, namespace, err)
	}

	var result []map[string]interface{}
	for _, deployment := range deployments.Items {
		deploymentInfo := map[string]interface{}{
			"name":              deployment.Name,
			"namespace":         deployment.Namespace,
			"replicas":          *deployment.Spec.Replicas,
			"readyReplicas":     deployment.Status.ReadyReplicas,
			"availableReplicas": deployment.Status.AvailableReplicas,
			"updatedReplicas":   deployment.Status.UpdatedReplicas,
			"creationTimestamp": deployment.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":            deployment.Labels,
			"selector":          deployment.Spec.Selector.MatchLabels,
			"strategy":          deployment.Spec.Strategy.Type,
		}
		result = append(result, deploymentInfo)
	}

	return result, nil
}

// GetDeployment returns detailed information about a specific deployment
func (c *Client) GetDeployment(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s' in namespace '%s': %v", name, namespace, err)
	}

	// Get replica sets
	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		fmt.Printf("Warning: failed to get replica sets: %v\n", err)
	}

	// Get pods
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		fmt.Printf("Warning: failed to get pods: %v\n", err)
	}

	result := map[string]interface{}{
		"name":                    deployment.Name,
		"namespace":               deployment.Namespace,
		"uid":                     deployment.UID,
		"resourceVersion":         deployment.ResourceVersion,
		"generation":              deployment.Generation,
		"creationTimestamp":       deployment.CreationTimestamp.Time.Format(time.RFC3339),
		"labels":                  deployment.Labels,
		"annotations":             deployment.Annotations,
		"replicas":                *deployment.Spec.Replicas,
		"selector":                deployment.Spec.Selector.MatchLabels,
		"strategy":                deployment.Spec.Strategy,
		"minReadySeconds":         deployment.Spec.MinReadySeconds,
		"progressDeadlineSeconds": deployment.Spec.ProgressDeadlineSeconds,
		"paused":                  deployment.Spec.Paused,
		"status": map[string]interface{}{
			"observedGeneration":  deployment.Status.ObservedGeneration,
			"replicas":            deployment.Status.Replicas,
			"updatedReplicas":     deployment.Status.UpdatedReplicas,
			"readyReplicas":       deployment.Status.ReadyReplicas,
			"availableReplicas":   deployment.Status.AvailableReplicas,
			"unavailableReplicas": deployment.Status.UnavailableReplicas,
			"conditions":          deployment.Status.Conditions,
		},
		"spec": map[string]interface{}{
			"template": deployment.Spec.Template,
		},
	}

	// Add replica set information
	if replicaSets != nil {
		var rsInfo []map[string]interface{}
		for _, rs := range replicaSets.Items {
			rsInfo = append(rsInfo, map[string]interface{}{
				"name":              rs.Name,
				"replicas":          *rs.Spec.Replicas,
				"readyReplicas":     rs.Status.ReadyReplicas,
				"availableReplicas": rs.Status.AvailableReplicas,
				"creationTimestamp": rs.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
		result["replicaSets"] = rsInfo
	}

	// Add pod information
	if pods != nil {
		var podInfo []map[string]interface{}
		for _, pod := range pods.Items {
			podInfo = append(podInfo, map[string]interface{}{
				"name":              pod.Name,
				"phase":             pod.Status.Phase,
				"ready":             isPodReady(&pod),
				"restarts":          getPodRestartCount(&pod),
				"creationTimestamp": pod.CreationTimestamp.Time.Format(time.RFC3339),
			})
		}
		result["pods"] = podInfo
	}

	return result, nil
}

// CreateDeployment creates a new deployment from a JSON manifest
func (c *Client) CreateDeployment(ctx context.Context, manifest string, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	var deployment appsv1.Deployment
	err := json.Unmarshal([]byte(manifest), &deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment manifest: %v", err)
	}

	// Ensure namespace is set
	deployment.Namespace = namespace

	// Set default values if not specified
	if deployment.Spec.Replicas == nil {
		replicas := int32(1)
		deployment.Spec.Replicas = &replicas
	}

	createdDeployment, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, &deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment '%s' in namespace '%s': %v", deployment.Name, namespace, err)
	}

	return createdDeployment, nil
}

// UpdateDeployment updates an existing deployment
func (c *Client) UpdateDeployment(ctx context.Context, name, manifest, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get existing deployment
	existingDeployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing deployment '%s': %v", name, err)
	}

	// Parse the updated manifest
	var updatedDeployment appsv1.Deployment
	err = json.Unmarshal([]byte(manifest), &updatedDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated deployment manifest: %v", err)
	}

	// Preserve important metadata
	updatedDeployment.Name = existingDeployment.Name
	updatedDeployment.Namespace = existingDeployment.Namespace
	updatedDeployment.ResourceVersion = existingDeployment.ResourceVersion
	updatedDeployment.UID = existingDeployment.UID

	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, &updatedDeployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment '%s' in namespace '%s': %v", name, namespace, err)
	}

	return result, nil
}

// DeleteDeployment deletes a deployment
func (c *Client) DeleteDeployment(ctx context.Context, name, namespace string, cascade bool) error {
	if namespace == "" {
		namespace = "default"
	}

	var propagationPolicy metav1.DeletionPropagation
	if cascade {
		propagationPolicy = metav1.DeletePropagationForeground
	} else {
		propagationPolicy = metav1.DeletePropagationOrphan
	}

	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete deployment '%s' in namespace '%s': %v", name, namespace, err)
	}

	return nil
}

// ScaleDeployment scales a deployment to the specified number of replicas
func (c *Client) ScaleDeployment(ctx context.Context, name, namespace string, replicas int32) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get the current deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Update the replica count
	deployment.Spec.Replicas = &replicas

	// Update the deployment
	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to scale deployment '%s' to %d replicas: %v", name, replicas, err)
	}

	return result, nil
}

// GetRolloutStatus returns the rollout status of a deployment
func (c *Client) GetRolloutStatus(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	status := map[string]interface{}{
		"name":                deployment.Name,
		"namespace":           deployment.Namespace,
		"generation":          deployment.Generation,
		"observedGeneration":  deployment.Status.ObservedGeneration,
		"replicas":            deployment.Status.Replicas,
		"updatedReplicas":     deployment.Status.UpdatedReplicas,
		"readyReplicas":       deployment.Status.ReadyReplicas,
		"availableReplicas":   deployment.Status.AvailableReplicas,
		"unavailableReplicas": deployment.Status.UnavailableReplicas,
		"conditions":          deployment.Status.Conditions,
		"paused":              deployment.Spec.Paused,
	}

	// Determine rollout status
	if deployment.Generation > deployment.Status.ObservedGeneration {
		status["rolloutStatus"] = "Waiting for rollout to finish"
	} else if deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
		status["rolloutStatus"] = "Waiting for deployment to update"
	} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		status["rolloutStatus"] = "Waiting for old replica sets to terminate"
	} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		status["rolloutStatus"] = "Waiting for deployment to become available"
	} else {
		status["rolloutStatus"] = "Successfully rolled out"
	}

	return status, nil
}

// GetRolloutHistory returns the rollout history of a deployment
func (c *Client) GetRolloutHistory(ctx context.Context, name, namespace string, revision *int64) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Get replica sets associated with this deployment
	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get replica sets: %v", err)
	}

	var history []map[string]interface{}
	for _, rs := range replicaSets.Items {
		// Get revision from annotation
		revisionStr, exists := rs.Annotations["deployment.kubernetes.io/revision"]
		if !exists {
			continue
		}

		// Parse revision number
		revisionNum, err := fmt.Sscanf(revisionStr, "%d")
		if err != nil {
			continue
		}

		// If specific revision requested, filter
		if revision != nil && int64(revisionNum) != *revision {
			continue
		}

		changeCase := rs.Annotations["deployment.kubernetes.io/revision-history-limit"]
		if changeCase == "" {
			changeCase = "No change cause specified"
		}

		historyEntry := map[string]interface{}{
			"revision":          revisionStr,
			"changeCause":       changeCase,
			"creationTimestamp": rs.CreationTimestamp.Time.Format(time.RFC3339),
			"replicaSetName":    rs.Name,
			"replicas":          *rs.Spec.Replicas,
			"template":          rs.Spec.Template,
		}

		history = append(history, historyEntry)
	}

	result := map[string]interface{}{
		"deployment": name,
		"namespace":  namespace,
		"history":    history,
	}

	return result, nil
}

// RollbackDeployment rolls back a deployment to a previous revision
func (c *Client) RollbackDeployment(ctx context.Context, name, namespace string, toRevision *int64) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Get replica sets to find the target revision
	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get replica sets: %v", err)
	}

	var targetRS *appsv1.ReplicaSet
	if toRevision != nil {
		// Find specific revision
		for _, rs := range replicaSets.Items {
			if revisionStr, exists := rs.Annotations["deployment.kubernetes.io/revision"]; exists {
				if revisionStr == fmt.Sprintf("%d", *toRevision) {
					targetRS = &rs
					break
				}
			}
		}
		if targetRS == nil {
			return nil, fmt.Errorf("revision %d not found", *toRevision)
		}
	} else {
		// Find previous revision (latest that's not current)
		currentRevision := deployment.Annotations["deployment.kubernetes.io/revision"]
		var latestRevision int64 = 0
		for _, rs := range replicaSets.Items {
			if revisionStr, exists := rs.Annotations["deployment.kubernetes.io/revision"]; exists && revisionStr != currentRevision {
				if rev, err := fmt.Sscanf(revisionStr, "%d"); err == nil && int64(rev) > latestRevision {
					latestRevision = int64(rev)
					targetRS = &rs
				}
			}
		}
		if targetRS == nil {
			return nil, fmt.Errorf("no previous revision found")
		}
	}

	// Update deployment template with target replica set template
	deployment.Spec.Template = targetRS.Spec.Template

	// Add rollback annotation
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["deployment.kubernetes.io/rollback-to"] = targetRS.Annotations["deployment.kubernetes.io/revision"]

	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to rollback deployment '%s': %v", name, err)
	}

	return result, nil
}

// PauseDeployment pauses a deployment
func (c *Client) PauseDeployment(ctx context.Context, name, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	deployment.Spec.Paused = true

	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pause deployment '%s': %v", name, err)
	}

	return result, nil
}

// ResumeDeployment resumes a paused deployment
func (c *Client) ResumeDeployment(ctx context.Context, name, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	deployment.Spec.Paused = false

	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to resume deployment '%s': %v", name, err)
	}

	return result, nil
}

// ========== EXTENDED DEPLOYMENT OPERATIONS ==========

// GetDeploymentEvents retrieves events related to a specific deployment
func (c *Client) GetDeploymentEvents(ctx context.Context, name, namespace string, limit int64) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	if limit <= 0 {
		limit = 50
	}

	// Verify deployment exists
	_, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Get events related to the deployment
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %v", err)
	}

	var result []map[string]interface{}
	for _, event := range events.Items {
		// Filter events related to this deployment or its resources
		if event.InvolvedObject.Kind == "Deployment" && event.InvolvedObject.Name == name ||
			event.InvolvedObject.Kind == "ReplicaSet" &&
				strings.HasPrefix(event.InvolvedObject.Name, name+"-") {

			eventInfo := map[string]interface{}{
				"type":           event.Type,
				"reason":         event.Reason,
				"message":        event.Message,
				"firstTimestamp": event.FirstTimestamp.Time.Format(time.RFC3339),
				"lastTimestamp":  event.LastTimestamp.Time.Format(time.RFC3339),
				"count":          event.Count,
				"involvedObject": map[string]interface{}{
					"kind": event.InvolvedObject.Kind,
					"name": event.InvolvedObject.Name,
				},
				"source": event.Source.Component,
			}
			result = append(result, eventInfo)
		}
	}

	return result, nil
}

// GetDeploymentLogs retrieves logs from all pods in a deployment
func (c *Client) GetDeploymentLogs(ctx context.Context, name, namespace, container string, lines int64, follow bool) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	if lines <= 0 {
		lines = 100
	}

	// Get deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Get pods for this deployment
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	result := map[string]interface{}{
		"deployment": name,
		"namespace":  namespace,
		"podLogs":    []map[string]interface{}{},
	}

	var podLogs []map[string]interface{}
	for _, pod := range pods.Items {
		podLogInfo := map[string]interface{}{
			"podName":    pod.Name,
			"containers": map[string]string{},
		}

		// Get containers to fetch logs from
		containers := []string{}
		if container != "" {
			containers = []string{container}
		} else {
			for _, c := range pod.Spec.Containers {
				containers = append(containers, c.Name)
			}
		}

		containerLogs := make(map[string]string)
		for _, containerName := range containers {
			req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
				Container: containerName,
				TailLines: &lines,
				Follow:    follow,
			})

			logs, err := req.Stream(ctx)
			if err != nil {
				containerLogs[containerName] = fmt.Sprintf("Error getting logs: %v", err)
				continue
			}
			defer logs.Close()

			buf := new(strings.Builder)
			_, err = io.Copy(buf, logs)
			if err != nil {
				containerLogs[containerName] = fmt.Sprintf("Error reading logs: %v", err)
			} else {
				containerLogs[containerName] = buf.String()
			}
		}

		podLogInfo["containers"] = containerLogs
		podLogs = append(podLogs, podLogInfo)
	}

	result["podLogs"] = podLogs
	return result, nil
}

// RestartDeployment restarts a deployment by triggering a rollout
func (c *Client) RestartDeployment(ctx context.Context, name, namespace string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get current deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Add restart annotation to trigger rollout
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// Update deployment
	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to restart deployment '%s': %v", name, err)
	}

	return result, nil
}

// WaitForDeployment waits for a deployment to reach its desired state
func (c *Client) WaitForDeployment(ctx context.Context, name, namespace string, timeoutSeconds int) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}

	timeout := time.Duration(timeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll deployment status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for deployment '%s' to be ready", name)
		case <-ticker.C:
			deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get deployment status: %v", err)
			}

			// Check if deployment is ready
			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
				deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas &&
				deployment.Status.ObservedGeneration >= deployment.Generation {

				return map[string]interface{}{
					"status":        "Ready",
					"message":       fmt.Sprintf("Deployment '%s' is ready with %d/%d replicas", name, deployment.Status.ReadyReplicas, *deployment.Spec.Replicas),
					"replicas":      *deployment.Spec.Replicas,
					"readyReplicas": deployment.Status.ReadyReplicas,
					"waitTime":      time.Since(time.Now().Add(-timeout)).String(),
				}, nil
			}
		}
	}
}

// SetDeploymentImage updates the image of a specific container in a deployment
func (c *Client) SetDeploymentImage(ctx context.Context, name, namespace, container, image string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get current deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Find and update the container image
	found := false
	for i, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			deployment.Spec.Template.Spec.Containers[i].Image = image
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("container '%s' not found in deployment '%s'", container, name)
	}

	// Update change cause annotation
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["deployment.kubernetes.io/change-cause"] = fmt.Sprintf("Updated image for container '%s' to '%s'", container, image)

	// Update deployment
	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment image: %v", err)
	}

	return result, nil
}

// SetDeploymentEnv updates environment variables in a deployment
func (c *Client) SetDeploymentEnv(ctx context.Context, name, namespace, container string, envVars map[string]string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get current deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Find and update the container environment variables
	found := false
	for i, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			// Create new env vars list
			envList := []corev1.EnvVar{}

			// Keep existing env vars that aren't being updated
			for _, existingEnv := range c.Env {
				if _, exists := envVars[existingEnv.Name]; !exists {
					envList = append(envList, existingEnv)
				}
			}

			// Add new/updated env vars
			for key, value := range envVars {
				envList = append(envList, corev1.EnvVar{
					Name:  key,
					Value: value,
				})
			}

			deployment.Spec.Template.Spec.Containers[i].Env = envList
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("container '%s' not found in deployment '%s'", container, name)
	}

	// Update change cause annotation
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["deployment.kubernetes.io/change-cause"] = fmt.Sprintf("Updated environment variables for container '%s'", container)

	// Update deployment
	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment environment: %v", err)
	}

	return result, nil
}

// PatchDeployment applies a patch to a deployment
func (c *Client) PatchDeployment(ctx context.Context, name, namespace string, patchData []byte, patchType types.PatchType) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	result, err := c.clientset.AppsV1().Deployments(namespace).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to patch deployment '%s': %v", name, err)
	}

	return result, nil
}

// GetDeploymentYAML exports a deployment as YAML
func (c *Client) GetDeploymentYAML(ctx context.Context, name, namespace string, export bool) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	if export {
		// Remove cluster-specific fields for export
		deployment.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		}
		deployment.ObjectMeta.UID = ""
		deployment.ObjectMeta.ResourceVersion = ""
		deployment.ObjectMeta.Generation = 0
		deployment.ObjectMeta.CreationTimestamp = metav1.Time{}
		deployment.ObjectMeta.SelfLink = ""
		deployment.ObjectMeta.ManagedFields = nil
		deployment.Status = appsv1.DeploymentStatus{}
	}

	yamlData, err := sigsyaml.Marshal(deployment)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment to YAML: %v", err)
	}

	return string(yamlData), nil
}

// SetDeploymentResources updates resource requests and limits
func (c *Client) SetDeploymentResources(ctx context.Context, name, namespace, container string, resources corev1.ResourceRequirements) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get current deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Find and update the container resources
	found := false
	for i, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			deployment.Spec.Template.Spec.Containers[i].Resources = resources
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("container '%s' not found in deployment '%s'", container, name)
	}

	// Update change cause annotation
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}
	deployment.Annotations["deployment.kubernetes.io/change-cause"] = fmt.Sprintf("Updated resources for container '%s'", container)

	// Update deployment
	result, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update deployment resources: %v", err)
	}

	return result, nil
}

// GetDeploymentMetrics gets CPU and memory metrics for a deployment
func (c *Client) GetDeploymentMetrics(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Note: This requires metrics-server to be installed in the cluster
	// For a basic implementation, we'll try to get pod metrics

	// Get deployment
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", name, err)
	}

	// Get pods for this deployment
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	result := map[string]interface{}{
		"deployment": name,
		"namespace":  namespace,
		"podCount":   len(pods.Items),
		"metrics":    "Metrics server integration required for detailed metrics",
		"pods":       []map[string]interface{}{},
	}

	// Basic pod resource information
	var podMetrics []map[string]interface{}
	for _, pod := range pods.Items {
		podInfo := map[string]interface{}{
			"name":  pod.Name,
			"phase": pod.Status.Phase,
			"ready": isPodReady(&pod),
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{},
				"limits":   map[string]interface{}{},
			},
		}

		// Get resource requests and limits from containers
		requests := make(map[string]interface{})
		limits := make(map[string]interface{})

		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				for resource, quantity := range container.Resources.Requests {
					requests[string(resource)] = quantity.String()
				}
			}
			if container.Resources.Limits != nil {
				for resource, quantity := range container.Resources.Limits {
					limits[string(resource)] = quantity.String()
				}
			}
		}

		podInfo["resources"].(map[string]interface{})["requests"] = requests
		podInfo["resources"].(map[string]interface{})["limits"] = limits

		podMetrics = append(podMetrics, podInfo)
	}

	result["pods"] = podMetrics
	return result, nil
}

// ListAllDeployments lists deployments across all namespaces
func (c *Client) ListAllDeployments(ctx context.Context, labelSelector string, includeSystem bool) (map[string]interface{}, error) {
	// Get all namespaces first
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
		"default":         false, // Include default namespace
	}

	result := map[string]interface{}{
		"totalDeployments": 0,
		"namespaces":       []map[string]interface{}{},
	}

	var allNamespaces []map[string]interface{}
	totalDeployments := 0

	for _, ns := range namespaces.Items {
		// Skip system namespaces if not requested
		if !includeSystem && systemNamespaces[ns.Name] {
			continue
		}

		deployments, err := c.clientset.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			continue // Skip this namespace if we can't list deployments
		}

		if len(deployments.Items) > 0 {
			nsInfo := map[string]interface{}{
				"namespace":       ns.Name,
				"deploymentCount": len(deployments.Items),
				"deployments":     []map[string]interface{}{},
			}

			var deploymentList []map[string]interface{}
			for _, deployment := range deployments.Items {
				deploymentInfo := map[string]interface{}{
					"name":              deployment.Name,
					"replicas":          *deployment.Spec.Replicas,
					"readyReplicas":     deployment.Status.ReadyReplicas,
					"availableReplicas": deployment.Status.AvailableReplicas,
					"creationTimestamp": deployment.CreationTimestamp.Time.Format(time.RFC3339),
					"labels":            deployment.Labels,
				}
				deploymentList = append(deploymentList, deploymentInfo)
			}

			nsInfo["deployments"] = deploymentList
			allNamespaces = append(allNamespaces, nsInfo)
			totalDeployments += len(deployments.Items)
		}
	}

	result["namespaces"] = allNamespaces
	result["totalDeployments"] = totalDeployments

	return result, nil
}

// ScaleAllDeployments scales all deployments in a namespace
func (c *Client) ScaleAllDeployments(ctx context.Context, namespace string, replicas int32, labelSelector string, dryRun bool) (map[string]interface{}, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace '%s': %v", namespace, err)
	}

	result := map[string]interface{}{
		"namespace":      namespace,
		"targetReplicas": replicas,
		"deployments":    []map[string]interface{}{},
		"dryRun":         dryRun,
		"totalProcessed": len(deployments.Items),
		"successful":     0,
		"failed":         0,
	}

	var deploymentResults []map[string]interface{}
	successful := 0
	failed := 0

	for _, deployment := range deployments.Items {
		deploymentResult := map[string]interface{}{
			"name":            deployment.Name,
			"currentReplicas": *deployment.Spec.Replicas,
			"targetReplicas":  replicas,
			"status":          "",
			"error":           "",
		}

		if !dryRun {
			// Update the deployment
			deployment.Spec.Replicas = &replicas
			_, err := c.clientset.AppsV1().Deployments(namespace).Update(ctx, &deployment, metav1.UpdateOptions{})
			if err != nil {
				deploymentResult["status"] = "failed"
				deploymentResult["error"] = err.Error()
				failed++
			} else {
				deploymentResult["status"] = "scaled"
				successful++
			}
		} else {
			deploymentResult["status"] = "dry-run"
			successful++
		}

		deploymentResults = append(deploymentResults, deploymentResult)
	}

	result["deployments"] = deploymentResults
	result["successful"] = successful
	result["failed"] = failed

	return result, nil
}

// ========== ADDITIONAL CLUSTER OVERVIEW OPERATIONS ==========

// GetNamespaceResourceUsage gets resource usage summary for a namespace
func (c *Client) GetNamespaceResourceUsage(ctx context.Context, namespace string, includeMetrics bool) (map[string]interface{}, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	// Get namespace info
	ns, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace '%s': %v", namespace, err)
	}

	result := map[string]interface{}{
		"namespace":         namespace,
		"creationTimestamp": ns.CreationTimestamp.Time.Format(time.RFC3339),
		"status":            ns.Status.Phase,
		"labels":            ns.Labels,
		"annotations":       ns.Annotations,
		"resourceCounts":    map[string]int{},
	}

	resourceCounts := make(map[string]interface{})

	// Count pods
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["pods"] = len(pods.Items)

		// Count pod phases
		podPhases := make(map[string]int)
		for _, pod := range pods.Items {
			podPhases[string(pod.Status.Phase)]++
		}
		resourceCounts["podPhases"] = podPhases
	}

	// Count deployments
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["deployments"] = len(deployments.Items)
	}

	// Count services
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["services"] = len(services.Items)
	}

	// Count configmaps
	configMaps, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["configMaps"] = len(configMaps.Items)
	}

	// Count secrets
	secrets, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["secrets"] = len(secrets.Items)
	}

	result["resourceCounts"] = resourceCounts
	return result, nil
}

// GetClusterOverview gets cluster-wide overview
func (c *Client) GetClusterOverview(ctx context.Context, includeMetrics bool) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"cluster": map[string]interface{}{
			"nodes":      map[string]interface{}{},
			"namespaces": map[string]interface{}{},
			"resources":  map[string]interface{}{},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Get nodes
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		nodeInfo := map[string]interface{}{
			"total": len(nodes.Items),
			"ready": 0,
			"nodes": []map[string]interface{}{},
		}

		var nodeList []map[string]interface{}
		readyNodes := 0
		for _, node := range nodes.Items {
			isReady := false
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					isReady = true
					readyNodes++
					break
				}
			}

			nodeDetails := map[string]interface{}{
				"name":              node.Name,
				"ready":             isReady,
				"creationTimestamp": node.CreationTimestamp.Time.Format(time.RFC3339),
				"labels":            node.Labels,
				"nodeInfo":          node.Status.NodeInfo,
			}
			nodeList = append(nodeList, nodeDetails)
		}

		nodeInfo["ready"] = readyNodes
		nodeInfo["nodes"] = nodeList
		result["cluster"].(map[string]interface{})["nodes"] = nodeInfo
	}

	// Get namespaces summary
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		nsInfo := map[string]interface{}{
			"total":      len(namespaces.Items),
			"active":     0,
			"namespaces": []map[string]interface{}{},
		}

		var nsList []map[string]interface{}
		activeNs := 0
		for _, ns := range namespaces.Items {
			if ns.Status.Phase == corev1.NamespaceActive {
				activeNs++
			}

			nsDetails := map[string]interface{}{
				"name":              ns.Name,
				"status":            ns.Status.Phase,
				"creationTimestamp": ns.CreationTimestamp.Time.Format(time.RFC3339),
				"labels":            ns.Labels,
			}
			nsList = append(nsList, nsDetails)
		}

		nsInfo["active"] = activeNs
		nsInfo["namespaces"] = nsList
		result["cluster"].(map[string]interface{})["namespaces"] = nsInfo
	}

	// Get cluster-wide resource counts
	resourceCounts := make(map[string]int)

	// Count all pods
	allPods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["totalPods"] = len(allPods.Items)
	}

	// Count all deployments
	allDeployments, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["totalDeployments"] = len(allDeployments.Items)
	}

	// Count all services
	allServices, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resourceCounts["totalServices"] = len(allServices.Items)
	}

	result["cluster"].(map[string]interface{})["resources"] = resourceCounts
	return result, nil
}

// ========== ADDITIONAL POD OPERATIONS ==========

// GetPodResourceUsage gets resource usage for a specific pod
func (c *Client) GetPodResourceUsage(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"pod":        name,
		"namespace":  namespace,
		"phase":      pod.Status.Phase,
		"containers": []map[string]interface{}{},
	}

	var containers []map[string]interface{}
	for _, container := range pod.Spec.Containers {
		containerInfo := map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{},
				"limits":   map[string]interface{}{},
			},
		}

		requests := make(map[string]interface{})
		limits := make(map[string]interface{})

		if container.Resources.Requests != nil {
			for resource, quantity := range container.Resources.Requests {
				requests[string(resource)] = quantity.String()
			}
		}

		if container.Resources.Limits != nil {
			for resource, quantity := range container.Resources.Limits {
				limits[string(resource)] = quantity.String()
			}
		}

		containerInfo["resources"].(map[string]interface{})["requests"] = requests
		containerInfo["resources"].(map[string]interface{})["limits"] = limits

		containers = append(containers, containerInfo)
	}

	result["containers"] = containers
	return result, nil
}

// GetPodsHealthStatus gets health status overview of pods in a namespace
func (c *Client) GetPodsHealthStatus(ctx context.Context, namespace, labelSelector string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	result := map[string]interface{}{
		"namespace": namespace,
		"totalPods": len(pods.Items),
		"summary":   map[string]int{},
		"pods":      []map[string]interface{}{},
	}

	summary := map[string]int{
		"Running":   0,
		"Pending":   0,
		"Succeeded": 0,
		"Failed":    0,
		"Unknown":   0,
		"Ready":     0,
		"NotReady":  0,
	}

	var podList []map[string]interface{}
	for _, pod := range pods.Items {
		phase := string(pod.Status.Phase)
		summary[phase]++

		isReady := isPodReady(&pod)
		if isReady {
			summary["Ready"]++
		} else {
			summary["NotReady"]++
		}

		podInfo := map[string]interface{}{
			"name":              pod.Name,
			"phase":             phase,
			"ready":             isReady,
			"restarts":          getPodRestartCount(&pod),
			"creationTimestamp": pod.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":            pod.Labels,
		}

		// Add container statuses
		var containerStatuses []map[string]interface{}
		for _, status := range pod.Status.ContainerStatuses {
			containerStatuses = append(containerStatuses, map[string]interface{}{
				"name":         status.Name,
				"ready":        status.Ready,
				"restartCount": status.RestartCount,
				"image":        status.Image,
			})
		}
		podInfo["containerStatuses"] = containerStatuses

		podList = append(podList, podInfo)
	}

	result["summary"] = summary
	result["pods"] = podList
	return result, nil
}

// ========== SERVICE OPERATIONS ==========

// ListServices returns a list of services in the specified namespace
func (c *Client) ListServices(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace '%s': %v", namespace, err)
	}

	var result []map[string]interface{}
	for _, service := range services.Items {
		serviceInfo := map[string]interface{}{
			"name":              service.Name,
			"namespace":         service.Namespace,
			"type":              string(service.Spec.Type),
			"clusterIP":         service.Spec.ClusterIP,
			"externalIPs":       service.Spec.ExternalIPs,
			"ports":             service.Spec.Ports,
			"selector":          service.Spec.Selector,
			"creationTimestamp": service.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":            service.Labels,
			"annotations":       service.Annotations,
		}

		// Add external access information
		if service.Spec.Type == corev1.ServiceTypeNodePort {
			serviceInfo["nodePort"] = service.Spec.Ports
		} else if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
			serviceInfo["loadBalancerIP"] = service.Spec.LoadBalancerIP
			serviceInfo["loadBalancerIngress"] = service.Status.LoadBalancer.Ingress
		}

		result = append(result, serviceInfo)
	}

	return result, nil
}

// ListServicesWithSelector returns services filtered by label selector
func (c *Client) ListServicesWithSelector(ctx context.Context, namespace, labelSelector string) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services with selector '%s' in namespace '%s': %v", labelSelector, namespace, err)
	}

	var result []map[string]interface{}
	for _, service := range services.Items {
		serviceInfo := map[string]interface{}{
			"name":              service.Name,
			"namespace":         service.Namespace,
			"type":              string(service.Spec.Type),
			"clusterIP":         service.Spec.ClusterIP,
			"ports":             service.Spec.Ports,
			"selector":          service.Spec.Selector,
			"creationTimestamp": service.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":            service.Labels,
		}
		result = append(result, serviceInfo)
	}

	return result, nil
}

// GetService returns detailed information about a specific service
func (c *Client) GetService(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service '%s' in namespace '%s': %v", name, namespace, err)
	}

	result := map[string]interface{}{
		"name":              service.Name,
		"namespace":         service.Namespace,
		"uid":               service.UID,
		"resourceVersion":   service.ResourceVersion,
		"creationTimestamp": service.CreationTimestamp.Time.Format(time.RFC3339),
		"labels":            service.Labels,
		"annotations":       service.Annotations,
		"spec": map[string]interface{}{
			"type":                     string(service.Spec.Type),
			"clusterIP":                service.Spec.ClusterIP,
			"clusterIPs":               service.Spec.ClusterIPs,
			"externalIPs":              service.Spec.ExternalIPs,
			"loadBalancerIP":           service.Spec.LoadBalancerIP,
			"loadBalancerSourceRanges": service.Spec.LoadBalancerSourceRanges,
			"externalName":             service.Spec.ExternalName,
			"externalTrafficPolicy":    service.Spec.ExternalTrafficPolicy,
			"healthCheckNodePort":      service.Spec.HealthCheckNodePort,
			"ports":                    service.Spec.Ports,
			"selector":                 service.Spec.Selector,
			"sessionAffinity":          service.Spec.SessionAffinity,
		},
		"status": service.Status,
	}

	// Get endpoints for this service
	endpoints, err := c.clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		result["endpoints"] = endpoints.Subsets
	}

	return result, nil
}

// CreateService creates a new service from a JSON manifest
func (c *Client) CreateService(ctx context.Context, manifest string, namespace string) (*corev1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}

	var service corev1.Service
	err := json.Unmarshal([]byte(manifest), &service)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service manifest: %v", err)
	}

	// Ensure namespace is set
	service.Namespace = namespace

	createdService, err := c.clientset.CoreV1().Services(namespace).Create(ctx, &service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service '%s' in namespace '%s': %v", service.Name, namespace, err)
	}

	return createdService, nil
}

// Fix the UpdateService method - add missing JSON parsing
func (c *Client) UpdateService(ctx context.Context, name, namespace, manifest string) (*corev1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}

	// First get the current service to get the resource version
	currentService, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get current service: %v", err)
	}

	var service corev1.Service
	err = json.Unmarshal([]byte(manifest), &service) // THIS LINE WAS MISSING
	if err != nil {
		return nil, fmt.Errorf("failed to parse service manifest: %v", err)
	}

	// Set the resource version and UID from current service
	service.ResourceVersion = currentService.ResourceVersion
	service.UID = currentService.UID
	service.Name = currentService.Name
	service.Namespace = currentService.Namespace

	result, err := c.clientset.CoreV1().Services(namespace).Update(ctx, &service, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update service '%s' in namespace '%s': %v", name, namespace, err)
	}

	return result, nil
}

// DeleteService deletes a service
func (c *Client) DeleteService(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}

	err := c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service '%s' in namespace '%s': %v", name, namespace, err)
	}

	return nil
}

// Improve the GetServiceEndpoints method
func (c *Client) GetServiceEndpoints(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get service first to verify it exists
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service '%s' in namespace '%s': %v", name, namespace, err)
	}

	// Get endpoints
	endpoints, err := c.clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// Handle missing endpoints gracefully
		if strings.Contains(err.Error(), "not found") {
			return map[string]interface{}{
				"serviceName": name,
				"namespace":   namespace,
				"serviceType": string(service.Spec.Type),
				"selector":    service.Spec.Selector,
				"endpoints":   nil,
				"ready":       false,
				"message":     "No endpoints found - service may not have ready pods matching the selector",
				"subsets":     []map[string]interface{}{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get endpoints for service '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"serviceName": name,
		"namespace":   namespace,
		"serviceType": string(service.Spec.Type),
		"selector":    service.Spec.Selector,
		"ready":       len(endpoints.Subsets) > 0,
		"subsets":     []map[string]interface{}{},
	}

	var subsets []map[string]interface{}
	for _, subset := range endpoints.Subsets {
		subsetInfo := map[string]interface{}{
			"addresses":         subset.Addresses,
			"notReadyAddresses": subset.NotReadyAddresses,
			"ports":             subset.Ports,
		}
		subsets = append(subsets, subsetInfo)
	}

	result["subsets"] = subsets
	return result, nil
}

// Improve TestServiceConnectivity method 
func (c *Client) TestServiceConnectivity(ctx context.Context, name, namespace string, port int32, protocol string) (map[string]interface{}, error) {
    if namespace == "" {
        namespace = "default"
    }
    if protocol == "" {
        protocol = "TCP"
    }

    // Get service
    service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get service '%s' in namespace '%s': %v", name, namespace, err)
    }

    // Try to get endpoints - handle gracefully if missing
    endpoints, err := c.clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
    hasEndpoints := err == nil && len(endpoints.Subsets) > 0

    result := map[string]interface{}{
        "serviceName":     name,
        "namespace":       namespace,
        "serviceType":     string(service.Spec.Type),
        "clusterIP":       service.Spec.ClusterIP,
        "hasEndpoints":    hasEndpoints,
        "connectivity":    map[string]interface{}{},
        "dnsNames":        []string{},
        "recommendations": []string{},
    }

    // DNS names for the service
    dnsNames := []string{
        name,
        fmt.Sprintf("%s.%s", name, namespace),
        fmt.Sprintf("%s.%s.svc", name, namespace),
        fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
    }
    result["dnsNames"] = dnsNames

    // Check connectivity
    connectivity := map[string]interface{}{
        "serviceExists":  true,
        "hasEndpoints":   hasEndpoints,
        "portAccessible": false,
        "dnsResolvable":  true,
    }

    // Validate port if specified
    if port > 0 {
        portFound := false
        for _, servicePort := range service.Spec.Ports {
            if servicePort.Port == port {
                portFound = true
                break
            }
        }
        connectivity["portAccessible"] = portFound
        if !portFound {
            result["recommendations"] = append(result["recommendations"].([]string), 
                fmt.Sprintf("Port %d not found in service ports", port))
        }
    }

    result["connectivity"] = connectivity

    // Add recommendations
    recommendations := result["recommendations"].([]string)
    if !hasEndpoints {
        recommendations = append(recommendations, 
            "Service has no endpoints - check if pods matching the selector are running and ready")
    }

    result["recommendations"] = recommendations
    return result, nil
}

// GetServiceEvents gets events related to a service
func (c *Client) GetServiceEvents(ctx context.Context, name, namespace string, limit int64) ([]map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}
	if limit <= 0 {
		limit = 50
	}

	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Service", name),
		Limit:         limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for service '%s': %v", name, err)
	}

	var result []map[string]interface{}
	for _, event := range events.Items {
		eventInfo := map[string]interface{}{
			"type":           event.Type,
			"reason":         event.Reason,
			"message":        event.Message,
			"firstTimestamp": event.FirstTimestamp.Time.Format(time.RFC3339),
			"lastTimestamp":  event.LastTimestamp.Time.Format(time.RFC3339),
			"count":          event.Count,
			"source":         event.Source.Component,
		}
		result = append(result, eventInfo)
	}

	return result, nil
}

// GetServiceYAML exports a service as YAML
func (c *Client) GetServiceYAML(ctx context.Context, name, namespace string, export bool) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get service '%s': %v", name, err)
	}

	if export {
		// Remove cluster-specific fields for export
		service.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		}
		service.ObjectMeta.UID = ""
		service.ObjectMeta.ResourceVersion = ""
		service.ObjectMeta.CreationTimestamp = metav1.Time{}
		service.ObjectMeta.SelfLink = ""
		service.ObjectMeta.ManagedFields = nil
		service.Spec.ClusterIP = ""
		service.Spec.ClusterIPs = nil
		service.Status = corev1.ServiceStatus{}
	}

	yamlData, err := sigsyaml.Marshal(service)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service to YAML: %v", err)
	}

	return string(yamlData), nil
}

// ExposeDeployment creates a service to expose a deployment
func (c *Client) ExposeDeployment(ctx context.Context, deploymentName, serviceName, namespace string, port, targetPort int32, serviceType string) (*corev1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}
	if serviceName == "" {
		serviceName = deploymentName
	}
	if targetPort == 0 {
		targetPort = port
	}
	if serviceType == "" {
		serviceType = "ClusterIP"
	}

	// Get deployment to extract selector
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment '%s': %v", deploymentName, err)
	}

	// Create service manifest
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       serviceName,
				"app.kubernetes.io/created-by": "k8s-mcp-server",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(serviceType),
			Selector: deployment.Spec.Selector.MatchLabels,
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	createdService, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service '%s': %v", serviceName, err)
	}

	return createdService, nil
}

// PatchService applies a patch to a service
func (c *Client) PatchService(ctx context.Context, name, namespace string, patchData []byte, patchType types.PatchType) (*corev1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}

	result, err := c.clientset.CoreV1().Services(namespace).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to patch service '%s': %v", name, err)
	}

	return result, nil
}

// ListAllServices lists services across all namespaces
func (c *Client) ListAllServices(ctx context.Context, labelSelector string, includeSystem bool) (map[string]interface{}, error) {
	// Get all namespaces first
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
		"default":         false, // Include default namespace
	}

	result := map[string]interface{}{
		"totalServices": 0,
		"namespaces":    []map[string]interface{}{},
	}

	var allNamespaces []map[string]interface{}
	totalServices := 0

	for _, ns := range namespaces.Items {
		// Skip system namespaces if not requested
		if !includeSystem && systemNamespaces[ns.Name] {
			continue
		}

		services, err := c.clientset.CoreV1().Services(ns.Name).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			continue // Skip this namespace if we can't list services
		}

		if len(services.Items) > 0 {
			nsInfo := map[string]interface{}{
				"namespace":    ns.Name,
				"serviceCount": len(services.Items),
				"services":     []map[string]interface{}{},
			}

			var serviceList []map[string]interface{}
			for _, service := range services.Items {
				serviceInfo := map[string]interface{}{
					"name":              service.Name,
					"type":              string(service.Spec.Type),
					"clusterIP":         service.Spec.ClusterIP,
					"ports":             service.Spec.Ports,
					"selector":          service.Spec.Selector,
					"creationTimestamp": service.CreationTimestamp.Time.Format(time.RFC3339),
					"labels":            service.Labels,
				}
				serviceList = append(serviceList, serviceInfo)
			}

			nsInfo["services"] = serviceList
			allNamespaces = append(allNamespaces, nsInfo)
			totalServices += len(services.Items)
		}
	}

	result["namespaces"] = allNamespaces
	result["totalServices"] = totalServices

	return result, nil
}

// GetServiceMetrics gets service metrics (basic implementation)
func (c *Client) GetServiceMetrics(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get service
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service '%s': %v", name, err)
	}

	// Get endpoints
	endpoints, err := c.clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %v", err)
	}

	result := map[string]interface{}{
		"serviceName": name,
		"namespace":   namespace,
		"serviceType": string(service.Spec.Type),
		"metrics": map[string]interface{}{
			"endpointCount":     0,
			"readyEndpoints":    0,
			"notReadyEndpoints": 0,
			"ports":             len(service.Spec.Ports),
		},
		"note": "For detailed traffic metrics, integrate with service mesh or monitoring solutions",
	}

	readyCount := 0
	notReadyCount := 0
	for _, subset := range endpoints.Subsets {
		readyCount += len(subset.Addresses)
		notReadyCount += len(subset.NotReadyAddresses)
	}

	result["metrics"].(map[string]interface{})["endpointCount"] = readyCount + notReadyCount
	result["metrics"].(map[string]interface{})["readyEndpoints"] = readyCount
	result["metrics"].(map[string]interface{})["notReadyEndpoints"] = notReadyCount

	return result, nil
}

// GetServiceTopology gets service topology information
func (c *Client) GetServiceTopology(ctx context.Context, name, namespace string) (map[string]interface{}, error) {
	if namespace == "" {
		namespace = "default"
	}

	// Get service
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service '%s': %v", name, err)
	}

	result := map[string]interface{}{
		"serviceName": name,
		"namespace":   namespace,
		"serviceType": string(service.Spec.Type),
		"selector":    service.Spec.Selector,
		"pods":        []map[string]interface{}{},
		"deployments": []map[string]interface{}{},
	}

	// Get pods that match the service selector
	if len(service.Spec.Selector) > 0 {
		labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: service.Spec.Selector,
		})

		pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			var podList []map[string]interface{}
			for _, pod := range pods.Items {
				podInfo := map[string]interface{}{
					"name":   pod.Name,
					"phase":  pod.Status.Phase,
					"ready":  isPodReady(&pod),
					"podIP":  pod.Status.PodIP,
					"labels": pod.Labels,
				}
				podList = append(podList, podInfo)
			}
			result["pods"] = podList
		}

		// Get deployments that might be controlling these pods
		deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			var deploymentList []map[string]interface{}
			for _, deployment := range deployments.Items {
				// Check if deployment selector matches service selector
				matches := true
				for key, value := range service.Spec.Selector {
					if deployment.Spec.Selector.MatchLabels[key] != value {
						matches = false
						break
					}
				}
				if matches {
					deploymentInfo := map[string]interface{}{
						"name":              deployment.Name,
						"replicas":          *deployment.Spec.Replicas,
						"readyReplicas":     deployment.Status.ReadyReplicas,
						"availableReplicas": deployment.Status.AvailableReplicas,
					}
					deploymentList = append(deploymentList, deploymentInfo)
				}
			}
			result["deployments"] = deploymentList
		}
	}

	return result, nil
}

// CreateServiceFromPods creates a service from pod selector
func (c *Client) CreateServiceFromPods(ctx context.Context, serviceName, namespace, labelSelector string, port, targetPort int32, serviceType string) (*corev1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}
	if targetPort == 0 {
		targetPort = port
	}
	if serviceType == "" {
		serviceType = "ClusterIP"
	}

	// Parse label selector
	selector, err := metav1.ParseToLabelSelector(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector '%s': %v", labelSelector, err)
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       serviceName,
				"app.kubernetes.io/created-by": "k8s-mcp-server",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(serviceType),
			Selector: selector.MatchLabels,
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	createdService, err := c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create service '%s': %v", serviceName, err)
	}

	return createdService, nil
}
