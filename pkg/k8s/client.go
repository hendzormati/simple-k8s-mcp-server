package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &Client{clientset: clientset}, nil
}

// TestConnection tests if we can connect to the cluster
func (c *Client) TestConnection() error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %v", err)
	}
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
