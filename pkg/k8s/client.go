package k8s

import (
	"context"
	"fmt"
	"path/filepath"

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

// GetPods returns all pods in the default namespace (for backwards compatibility)
func (c *Client) GetPods() ([]string, error) {
	return c.GetPodsInNamespace("default")
}

// GetPodsInNamespace returns all pods in the specified namespace
func (c *Client) GetPodsInNamespace(namespace string) ([]string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
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

// ForceDeleteNamespace attempts to force delete a namespace by removing finalizers
func (c *Client) ForceDeleteNamespace(ctx context.Context, name string) error {
	// First, try to get the namespace to see its current state
	namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get namespace '%s': %v", name, err)
	}

	// If it's already in Terminating state and has finalizers, remove them
	if namespace.Status.Phase == corev1.NamespaceTerminating && len(namespace.Spec.Finalizers) > 0 {
		// Clear the finalizers
		namespace.Spec.Finalizers = []corev1.FinalizerName{}

		// Update the namespace
		_, err = c.clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove finalizers from namespace '%s': %v", name, err)
		}

		return nil
	}

	// If it's not terminating, try regular delete
	return c.DeleteNamespace(ctx, name)
}
