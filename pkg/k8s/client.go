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

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				fmt.Printf("🐄 Found K3s kubeconfig at: %s\n", k3sPath)
				config, err = clientcmd.BuildConfigFromFlags("", k3sPath)
				if err == nil {
					configSource = fmt.Sprintf("K3s config (%s)", k3sPath)
					fmt.Printf("✅ Successfully loaded K3s configuration\n")
					break
				} else {
					fmt.Printf("⚠️  Failed to load K3s config from %s: %v\n", k3sPath, err)
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
