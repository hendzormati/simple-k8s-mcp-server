package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Helper function to safely get arguments as map
func getArguments(request mcp.CallToolRequest) map[string]interface{} {
	if request.Params.Arguments == nil {
		return make(map[string]interface{})
	}

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		return args
	}

	return make(map[string]interface{})
}

// Helper function to parse JSON string to map[string]string
func parseJSONStringToMap(jsonStr string) (map[string]string, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var result map[string]string
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	return result, nil
}

// ========== NAMESPACE HANDLERS ==========

// ListNamespaces returns a handler function for the listNamespaces tool
func ListNamespaces(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// List namespaces
		namespaces, err := client.ListNamespaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"namespaces": namespaces,
			"count":      len(namespaces),
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespace returns a handler function for the getNamespace tool
func GetNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace details
		namespace, err := client.GetNamespace(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace: %v", err)
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// CreateNamespace returns a handler function for the createNamespace tool
func CreateNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get optional labels (parse from JSON string)
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsStr, ok := labelsArg.(string); ok && labelsStr != "" {
				parsedLabels, err := parseJSONStringToMap(labelsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid labels JSON: %v", err)
				}
				labels = parsedLabels
			}
		}

		// Get optional annotations (parse from JSON string)
		var annotations map[string]string
		if annotationsArg, exists := args["annotations"]; exists {
			if annotationsStr, ok := annotationsArg.(string); ok && annotationsStr != "" {
				parsedAnnotations, err := parseJSONStringToMap(annotationsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid annotations JSON: %v", err)
				}
				annotations = parsedAnnotations
			}
		}

		// Create namespace
		namespace, err := client.CreateNamespace(ctx, nameStr, labels, annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace: %v", err)
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// UpdateNamespace returns a handler function for the updateNamespace tool
func UpdateNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get optional labels (parse from JSON string)
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsStr, ok := labelsArg.(string); ok && labelsStr != "" {
				parsedLabels, err := parseJSONStringToMap(labelsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid labels JSON: %v", err)
				}
				labels = parsedLabels
			}
		}

		// Get optional annotations (parse from JSON string)
		var annotations map[string]string
		if annotationsArg, exists := args["annotations"]; exists {
			if annotationsStr, ok := annotationsArg.(string); ok && annotationsStr != "" {
				parsedAnnotations, err := parseJSONStringToMap(annotationsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid annotations JSON: %v", err)
				}
				annotations = parsedAnnotations
			}
		}

		// Update namespace
		namespace, err := client.UpdateNamespace(ctx, nameStr, labels, annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to update namespace: %v", err)
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// DeleteNamespace returns a handler function for the deleteNamespace tool
func DeleteNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Check if namespace exists and get its current state
		namespace, err := client.GetNamespace(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("namespace '%s' not found: %v", nameStr, err)
		}

		// Check for resources in namespace
		hasResources := false
		if namespace != nil {
			// You could add a check here to warn about resources
			_ = namespace // Placeholder for future resource checking
		}

		// Attempt deletion
		err = client.DeleteNamespace(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to delete namespace: %v", err)
		}

		// Wait a moment and check if it's actually deleting
		time.Sleep(2 * time.Second)

		// Check if namespace is in terminating state
		updatedNs, err := client.GetNamespace(ctx, nameStr)
		var status string = "deleted"
		var message string = fmt.Sprintf("Namespace '%s' deleted successfully", nameStr)

		if err == nil {
			// Namespace still exists, check its status
			nsMap := updatedNs
			if statusVal, exists := nsMap["status"]; exists {
				if statusStr, ok := statusVal.(string); ok && statusStr == "Terminating" {
					status = "terminating"
					message = fmt.Sprintf("Namespace '%s' is terminating. If it gets stuck, use forceDeleteNamespace", nameStr)

					// Add helpful information about what might be blocking
					if hasResources {
						message += " (contains resources that may delay deletion)"
					}
				}
			}
		}

		response := map[string]interface{}{
			"message":   message,
			"namespace": nameStr,
			"status":    status,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SmartDeleteNamespace returns a handler that automatically chooses the best deletion strategy
func SmartDeleteNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get force flag (default true)
		force := true
		if forceArg, exists := args["force"]; exists {
			if forceBool, ok := forceArg.(bool); ok {
				force = forceBool
			}
		}

		// Try regular delete first
		err := client.DeleteNamespace(ctx, nameStr)
		if err != nil {
			if force {
				// If regular delete fails and force is enabled, try force delete
				err = client.ForceDeleteNamespace(ctx, nameStr)
				if err != nil {
					return nil, fmt.Errorf("failed to delete namespace (tried regular and force): %v", err)
				}

				response := map[string]interface{}{
					"message":   fmt.Sprintf("Namespace '%s' force deleted successfully", nameStr),
					"namespace": nameStr,
					"status":    "force-deleted",
					"method":    "enhanced-force-delete",
				}

				jsonResponse, err := json.Marshal(response)
				if err != nil {
					return nil, fmt.Errorf("failed to serialize response: %v", err)
				}

				return mcp.NewToolResultText(string(jsonResponse)), nil
			} else {
				return nil, fmt.Errorf("failed to delete namespace: %v", err)
			}
		}

		// Regular delete succeeded
		response := map[string]interface{}{
			"message":   fmt.Sprintf("Namespace '%s' deleted successfully", nameStr),
			"namespace": nameStr,
			"status":    "deleted",
			"method":    "regular-delete",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespaceResourceQuota returns a handler function for the getNamespaceResourceQuota tool
func GetNamespaceResourceQuota(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get resource quotas
		quotas, err := client.GetNamespaceResourceQuota(ctx, namespaceStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource quotas: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"namespace":      namespaceStr,
			"resourceQuotas": quotas,
			"count":          len(quotas),
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// Add these handlers to your k8s.go file in the handlers package

// GetNamespaceEvents returns a handler function for the getNamespaceEvents tool
func GetNamespaceEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get events
		events, err := client.GetNamespaceEvents(ctx, namespaceStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get events: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"namespace": namespaceStr,
			"events":    events,
			"count":     len(events),
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespaceAllResources returns a handler function for the getNamespaceAllResources tool
func GetNamespaceAllResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get all resources
		resources, err := client.GetNamespaceAllResources(ctx, namespaceStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources: %v", err)
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ForceDeleteNamespace returns a handler function for the forceDeleteNamespace tool
func ForceDeleteNamespace(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := getArguments(request)
		if len(args) == 0 {
			return nil, fmt.Errorf("missing arguments")
		}

		// Get namespace name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Force delete namespace
		err := client.ForceDeleteNamespace(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to force delete namespace: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"message":   fmt.Sprintf("Namespace '%s' force deletion initiated (finalizers removed if needed)", nameStr),
			"namespace": nameStr,
			"status":    "force-deleted",
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespaceYAML returns a handler function for the getNamespaceYAML tool
func GetNamespaceYAML(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		yamlDef, err := client.GetNamespaceYAML(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace YAML: %v", err)
		}

		response := map[string]interface{}{
			"name": nameStr,
			"yaml": yamlDef,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SetNamespaceResourceQuota returns a handler function for the setNamespaceResourceQuota tool
func SetNamespaceResourceQuota(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		manifest, exists := args["manifest"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: manifest")
		}
		manifestStr, ok := manifest.(string)
		if !ok || manifestStr == "" {
			return nil, fmt.Errorf("manifest must be a non-empty string")
		}

		quota, err := client.SetNamespaceResourceQuota(ctx, namespaceStr, manifestStr)
		if err != nil {
			return nil, fmt.Errorf("failed to set resource quota: %v", err)
		}

		response := map[string]interface{}{
			"message":       fmt.Sprintf("Resource quota '%s' %s successfully in namespace '%s'", quota["name"], quota["operation"], namespaceStr),
			"namespace":     namespaceStr,
			"resourceQuota": quota,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespaceLimitRanges returns a handler function for the getNamespaceLimitRanges tool
func GetNamespaceLimitRanges(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		limitRanges, err := client.GetNamespaceLimitRanges(ctx, namespaceStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get limit ranges: %v", err)
		}

		response := map[string]interface{}{
			"namespace":   namespaceStr,
			"limitRanges": limitRanges,
			"count":       len(limitRanges),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SetNamespaceLimitRange returns a handler function for the setNamespaceLimitRange tool
func SetNamespaceLimitRange(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		manifest, exists := args["manifest"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: manifest")
		}
		manifestStr, ok := manifest.(string)
		if !ok || manifestStr == "" {
			return nil, fmt.Errorf("manifest must be a non-empty string")
		}

		limitRange, err := client.SetNamespaceLimitRange(ctx, namespaceStr, manifestStr)
		if err != nil {
			return nil, fmt.Errorf("failed to set limit range: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Limit range '%s' %s successfully in namespace '%s'", limitRange["name"], limitRange["operation"], namespaceStr),
			"namespace":  namespaceStr,
			"limitRange": limitRange,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ========== POD HANDLERS ==========

// ListPods returns a handler function for the listPods tool
func ListPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		var pods []map[string]interface{}
		var err error

		if labelSelector, exists := args["labelSelector"]; exists {
			if selectorStr, ok := labelSelector.(string); ok && selectorStr != "" {
				pods, err = client.GetPodsInNamespaceWithSelector(namespace, selectorStr)
			} else {
				pods, err = client.GetPodsInNamespace(namespace)
			}
		} else {
			pods, err = client.GetPodsInNamespace(namespace)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get pods: %v", err)
		}

		response := map[string]interface{}{
			"namespace": namespace,
			"pods":      pods,
			"count":     len(pods),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPod returns a handler function for the getPod tool
func GetPod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		pod, err := client.GetPod(ctx, namespaceStr, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod: %v", err)
		}

		jsonResponse, err := json.Marshal(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodLogs returns a handler function for the getPodLogs tool
func GetPodLogs(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Required arguments
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Optional arguments
		containerName := ""
		if container, exists := args["containerName"]; exists {
			if containerStr, ok := container.(string); ok {
				containerName = containerStr
			}
		}

		tailLines := int64(100)
		if lines, exists := args["tailLines"]; exists {
			if linesFloat, ok := lines.(float64); ok {
				tailLines = int64(linesFloat)
			}
		}

		follow := false
		if followArg, exists := args["follow"]; exists {
			if followBool, ok := followArg.(bool); ok {
				follow = followBool
			}
		}

		previous := false
		if prevArg, exists := args["previous"]; exists {
			if prevBool, ok := prevArg.(bool); ok {
				previous = prevBool
			}
		}

		logs, err := client.GetPodLogs(ctx, namespaceStr, nameStr, containerName, tailLines, follow, previous)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod logs: %v", err)
		}

		response := map[string]interface{}{
			"podName":       nameStr,
			"namespace":     namespaceStr,
			"containerName": containerName,
			"logs":          logs,
			"tailLines":     tailLines,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// DeletePod returns a handler function for the deletePod tool
func DeletePod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		gracePeriodSeconds := int64(30)
		if grace, exists := args["gracePeriodSeconds"]; exists {
			if graceFloat, ok := grace.(float64); ok {
				gracePeriodSeconds = int64(graceFloat)
			}
		}

		err := client.DeletePod(ctx, namespaceStr, nameStr, gracePeriodSeconds)
		if err != nil {
			return nil, fmt.Errorf("failed to delete pod: %v", err)
		}

		response := map[string]interface{}{
			"message":            fmt.Sprintf("Pod '%s' in namespace '%s' deleted successfully", nameStr, namespaceStr),
			"podName":            nameStr,
			"namespace":          namespaceStr,
			"gracePeriodSeconds": gracePeriodSeconds,
			"status":             "deleted",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodEvents returns a handler function for the getPodEvents tool
func GetPodEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		events, err := client.GetPodEvents(ctx, namespaceStr, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod events: %v", err)
		}

		response := map[string]interface{}{
			"podName":   nameStr,
			"namespace": namespaceStr,
			"events":    events,
			"count":     len(events),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RestartPod returns a handler function for the restartPod tool
func RestartPod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Delete the pod with grace period of 0 for immediate restart
		err := client.DeletePod(ctx, namespaceStr, nameStr, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to restart pod: %v", err)
		}

		response := map[string]interface{}{
			"message":   fmt.Sprintf("Pod '%s' in namespace '%s' restarted successfully", nameStr, namespaceStr),
			"podName":   nameStr,
			"namespace": namespaceStr,
			"status":    "restarted",
			"note":      "Pod has been deleted and will be recreated by its controller (if managed by a Deployment, ReplicaSet, etc.)",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// DescribePod returns a handler function for the describePod tool
func DescribePod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get detailed pod information
		pod, err := client.GetPod(ctx, namespaceStr, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod: %v", err)
		}

		// Get pod events
		events, err := client.GetPodEvents(ctx, namespaceStr, nameStr)
		if err != nil {
			// Don't fail if events can't be retrieved, just log it
			events = []map[string]interface{}{}
		}

		// Combine pod details with events for a comprehensive description
		response := map[string]interface{}{
			"podDetails": pod,
			"events":     events,
			"summary": map[string]interface{}{
				"name":      nameStr,
				"namespace": namespaceStr,
				"status":    pod["status"],
				"ready":     pod["ready"],
				"restarts":  pod["restartCount"],
				"age":       pod["creationTimestamp"],
				"node":      pod["nodeName"],
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodMetrics returns a handler function for the getPodMetrics tool
func GetPodMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Note: For now, we'll return resource requests/limits from the pod spec
		// To get actual metrics, you would need metrics-server installed and use metrics API
		pod, err := client.GetPod(ctx, namespaceStr, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod: %v", err)
		}

		// Extract resource information from containers
		containers, ok := pod["containers"].([]map[string]interface{})
		if !ok {
			containers = []map[string]interface{}{}
		}

		response := map[string]interface{}{
			"podName":    nameStr,
			"namespace":  namespaceStr,
			"status":     pod["status"],
			"containers": containers,
			"note":       "Resource requests/limits shown. For actual usage metrics, ensure metrics-server is installed in your cluster.",
			"metrics": map[string]interface{}{
				"available": false,
				"reason":    "Metrics collection requires metrics-server to be installed and configured",
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// CreatePod returns a handler function for the createPod tool
func CreatePod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get required namespace
		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get required manifest
		manifest, exists := args["manifest"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: manifest")
		}
		manifestStr, ok := manifest.(string)
		if !ok || manifestStr == "" {
			return nil, fmt.Errorf("manifest must be a non-empty string")
		}

		// Create the pod
		pod, err := client.CreatePod(ctx, namespaceStr, manifestStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create pod: %v", err)
		}

		response := map[string]interface{}{
			"message": fmt.Sprintf("Pod '%s' created successfully in namespace '%s'", pod["name"], namespaceStr),
			"pod":     pod,
			"status":  "created",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// UpdatePod returns a handler function for the updatePod tool
func UpdatePod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get required arguments
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		// Get optional labels (parse from JSON string)
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsStr, ok := labelsArg.(string); ok && labelsStr != "" {
				parsedLabels, err := parseJSONStringToMap(labelsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid labels JSON: %v", err)
				}
				labels = parsedLabels
			}
		}

		// Get optional annotations (parse from JSON string)
		var annotations map[string]string
		if annotationsArg, exists := args["annotations"]; exists {
			if annotationsStr, ok := annotationsArg.(string); ok && annotationsStr != "" {
				parsedAnnotations, err := parseJSONStringToMap(annotationsStr)
				if err != nil {
					return nil, fmt.Errorf("invalid annotations JSON: %v", err)
				}
				annotations = parsedAnnotations
			}
		}

		// Check if at least one update is provided
		if labels == nil && annotations == nil {
			return nil, fmt.Errorf("at least one of 'labels' or 'annotations' must be provided")
		}

		// Update the pod
		pod, err := client.UpdatePod(ctx, namespaceStr, nameStr, labels, annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to update pod: %v", err)
		}

		response := map[string]interface{}{
			"message": fmt.Sprintf("Pod '%s' in namespace '%s' updated successfully", nameStr, namespaceStr),
			"pod":     pod,
			"status":  "updated",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ========== DEPLOYMENT HANDLERS ==========

// ListDeployments returns a handler function for the listDeployments tool
func ListDeployments(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		// Check for label selector
		var deployments []map[string]interface{}
		var err error

		if labelSelector, exists := args["labelSelector"]; exists {
			if selectorStr, ok := labelSelector.(string); ok && selectorStr != "" {
				deployments, err = client.ListDeploymentsWithSelector(ctx, namespace, selectorStr)
			} else {
				deployments, err = client.ListDeployments(ctx, namespace)
			}
		} else {
			deployments, err = client.ListDeployments(ctx, namespace)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list deployments: %v", err)
		}

		response := map[string]interface{}{
			"deployments": deployments,
			"namespace":   namespace,
			"count":       len(deployments),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetDeployment returns a handler function for the getDeployment tool
func GetDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.GetDeployment(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment: %v", err)
		}

		jsonResponse, err := json.Marshal(deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// CreateDeployment returns a handler function for the createDeployment tool
func CreateDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get manifest
		manifest, exists := args["manifest"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: manifest")
		}
		manifestStr, ok := manifest.(string)
		if !ok || manifestStr == "" {
			return nil, fmt.Errorf("manifest must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.CreateDeployment(ctx, manifestStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to create deployment: %v", err)
		}

		response := map[string]interface{}{
			"message": fmt.Sprintf("Deployment '%s' created successfully in namespace '%s'", deployment.Name, deployment.Namespace),
			"deployment": map[string]interface{}{
				"name":              deployment.Name,
				"namespace":         deployment.Namespace,
				"uid":               deployment.UID,
				"replicas":          *deployment.Spec.Replicas,
				"creationTimestamp": deployment.CreationTimestamp.Time.Format(time.RFC3339),
				"labels":            deployment.Labels,
				"selector":          deployment.Spec.Selector.MatchLabels,
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// UpdateDeployment returns a handler function for the updateDeployment tool
func UpdateDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get manifest
		manifest, exists := args["manifest"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: manifest")
		}
		manifestStr, ok := manifest.(string)
		if !ok || manifestStr == "" {
			return nil, fmt.Errorf("manifest must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.UpdateDeployment(ctx, nameStr, manifestStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to update deployment: %v", err)
		}

		response := map[string]interface{}{
			"message": fmt.Sprintf("Deployment '%s' updated successfully", deployment.Name),
			"deployment": map[string]interface{}{
				"name":             deployment.Name,
				"namespace":        deployment.Namespace,
				"generation":       deployment.Generation,
				"replicas":         *deployment.Spec.Replicas,
				"updatedTimestamp": deployment.Status.Conditions,
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// DeleteDeployment returns a handler function for the deleteDeployment tool
func DeleteDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		// Get cascade option (default to true)
		cascade := true
		if cascadeArg, exists := args["cascade"]; exists {
			if cascadeBool, ok := cascadeArg.(bool); ok {
				cascade = cascadeBool
			}
		}

		err := client.DeleteDeployment(ctx, nameStr, namespace, cascade)
		if err != nil {
			return nil, fmt.Errorf("failed to delete deployment: %v", err)
		}

		cascadeStr := "with cascade (includes replica sets and pods)"
		if !cascade {
			cascadeStr = "without cascade (orphaning replica sets and pods)"
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' deleted successfully %s", nameStr, cascadeStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"cascade":    cascade,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ScaleDeployment returns a handler function for the scaleDeployment tool
func ScaleDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get replicas
		replicas, exists := args["replicas"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: replicas")
		}

		var replicasInt32 int32
		switch v := replicas.(type) {
		case float64:
			replicasInt32 = int32(v)
		case int:
			replicasInt32 = int32(v)
		case int32:
			replicasInt32 = v
		default:
			return nil, fmt.Errorf("replicas must be a number")
		}

		if replicasInt32 < 0 {
			return nil, fmt.Errorf("replicas must be non-negative")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.ScaleDeployment(ctx, nameStr, namespace, replicasInt32)
		if err != nil {
			return nil, fmt.Errorf("failed to scale deployment: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' scaled to %d replicas", nameStr, replicasInt32),
			"deployment": nameStr,
			"namespace":  namespace,
			"replicas":   replicasInt32,
			"status": map[string]interface{}{
				"currentReplicas": deployment.Status.Replicas,
				"updatedReplicas": deployment.Status.UpdatedReplicas,
				"readyReplicas":   deployment.Status.ReadyReplicas,
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RolloutStatus returns a handler function for the rolloutStatus tool
func RolloutStatus(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		status, err := client.GetRolloutStatus(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get rollout status: %v", err)
		}

		jsonResponse, err := json.Marshal(status)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RolloutHistory returns a handler function for the rolloutHistory tool
func RolloutHistory(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		// Get optional revision
		var revision *int64
		if rev, exists := args["revision"]; exists {
			switch v := rev.(type) {
			case float64:
				revInt := int64(v)
				revision = &revInt
			case int:
				revInt := int64(v)
				revision = &revInt
			case int64:
				revision = &v
			}
		}

		history, err := client.GetRolloutHistory(ctx, nameStr, namespace, revision)
		if err != nil {
			return nil, fmt.Errorf("failed to get rollout history: %v", err)
		}

		jsonResponse, err := json.Marshal(history)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RolloutUndo returns a handler function for the rolloutUndo tool
func RolloutUndo(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		// Get optional toRevision
		var toRevision *int64
		if rev, exists := args["toRevision"]; exists {
			switch v := rev.(type) {
			case float64:
				revInt := int64(v)
				toRevision = &revInt
			case int:
				revInt := int64(v)
				toRevision = &revInt
			case int64:
				toRevision = &v
			}
		}

		deployment, err := client.RollbackDeployment(ctx, nameStr, namespace, toRevision)
		if err != nil {
			return nil, fmt.Errorf("failed to rollback deployment: %v", err)
		}

		var revisionStr string
		if toRevision != nil {
			revisionStr = fmt.Sprintf(" to revision %d", *toRevision)
		} else {
			revisionStr = " to previous revision"
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' rolled back%s successfully", nameStr, revisionStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"generation": deployment.Generation,
			"toRevision": toRevision,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// PauseDeployment returns a handler function for the pauseDeployment tool
func PauseDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.PauseDeployment(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to pause deployment: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' paused successfully", nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"paused":     deployment.Spec.Paused,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ResumeDeployment returns a handler function for the resumeDeployment tool
func ResumeDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		// Get deployment name
		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		// Get namespace (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.ResumeDeployment(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to resume deployment: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' resumed successfully", nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"paused":     deployment.Spec.Paused,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ========== EXTENDED DEPLOYMENT HANDLERS ==========

// GetDeploymentEvents returns a handler function for the getDeploymentEvents tool
func GetDeploymentEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		limit := int64(50)
		if limitArg, exists := args["limit"]; exists {
			switch v := limitArg.(type) {
			case float64:
				limit = int64(v)
			case int:
				limit = int64(v)
			case int64:
				limit = v
			}
		}

		events, err := client.GetDeploymentEvents(ctx, nameStr, namespace, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment events: %v", err)
		}

		response := map[string]interface{}{
			"deployment": nameStr,
			"namespace":  namespace,
			"events":     events,
			"count":      len(events),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetDeploymentLogs returns a handler function for the getDeploymentLogs tool
func GetDeploymentLogs(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		container := ""
		if containerArg, exists := args["container"]; exists {
			if containerStr, ok := containerArg.(string); ok {
				container = containerStr
			}
		}

		lines := int64(100)
		if linesArg, exists := args["lines"]; exists {
			switch v := linesArg.(type) {
			case float64:
				lines = int64(v)
			case int:
				lines = int64(v)
			case int64:
				lines = v
			}
		}

		follow := false
		if followArg, exists := args["follow"]; exists {
			if followBool, ok := followArg.(bool); ok {
				follow = followBool
			}
		}

		logs, err := client.GetDeploymentLogs(ctx, nameStr, namespace, container, lines, follow)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment logs: %v", err)
		}

		jsonResponse, err := json.Marshal(logs)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RestartDeployment returns a handler function for the restartDeployment tool
func RestartDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.RestartDeployment(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to restart deployment: %v", err)
		}

		response := map[string]interface{}{
			"message":     fmt.Sprintf("Deployment '%s' restarted successfully", nameStr),
			"deployment":  nameStr,
			"namespace":   namespace,
			"restartedAt": deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"],
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// WaitForDeployment returns a handler function for the waitForDeployment tool
func WaitForDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		timeout := 300
		if timeoutArg, exists := args["timeout"]; exists {
			switch v := timeoutArg.(type) {
			case float64:
				timeout = int(v)
			case int:
				timeout = v
			}
		}

		result, err := client.WaitForDeployment(ctx, nameStr, namespace, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for deployment: %v", err)
		}

		jsonResponse, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SetDeploymentImage returns a handler function for the setDeploymentImage tool
func SetDeploymentImage(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		container, exists := args["container"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: container")
		}
		containerStr, ok := container.(string)
		if !ok || containerStr == "" {
			return nil, fmt.Errorf("container must be a non-empty string")
		}

		image, exists := args["image"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: image")
		}
		imageStr, ok := image.(string)
		if !ok || imageStr == "" {
			return nil, fmt.Errorf("image must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.SetDeploymentImage(ctx, nameStr, namespace, containerStr, imageStr)
		if err != nil {
			return nil, fmt.Errorf("failed to set deployment image: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Image updated to '%s' for container '%s' in deployment '%s'", imageStr, containerStr, nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"container":  containerStr,
			"image":      imageStr,
			"generation": deployment.Generation,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SetDeploymentEnv returns a handler function for the setDeploymentEnv tool
func SetDeploymentEnv(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		container, exists := args["container"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: container")
		}
		containerStr, ok := container.(string)
		if !ok || containerStr == "" {
			return nil, fmt.Errorf("container must be a non-empty string")
		}

		env, exists := args["env"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: env")
		}
		envStr, ok := env.(string)
		if !ok || envStr == "" {
			return nil, fmt.Errorf("env must be a non-empty string")
		}

		// Parse environment variables JSON
		var envVars map[string]string
		err := json.Unmarshal([]byte(envStr), &envVars)
		if err != nil {
			return nil, fmt.Errorf("invalid env JSON: %v", err)
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.SetDeploymentEnv(ctx, nameStr, namespace, containerStr, envVars)
		if err != nil {
			return nil, fmt.Errorf("failed to set deployment environment variables: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Environment variables updated for container '%s' in deployment '%s'", containerStr, nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"container":  containerStr,
			"envVars":    envVars,
			"generation": deployment.Generation,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ========== ADDITIONAL POD HANDLERS ==========

// GetPodResourceUsage returns a handler function for the getPodResourceUsage tool
func GetPodResourceUsage(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		usage, err := client.GetPodResourceUsage(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod resource usage: %v", err)
		}

		jsonResponse, err := json.Marshal(usage)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodsHealthStatus returns a handler function for the getPodsHealthStatus tool
func GetPodsHealthStatus(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		labelSelector := ""
		if selector, exists := args["labelSelector"]; exists {
			if selectorStr, ok := selector.(string); ok {
				labelSelector = selectorStr
			}
		}

		healthStatus, err := client.GetPodsHealthStatus(ctx, namespace, labelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods health status: %v", err)
		}

		jsonResponse, err := json.Marshal(healthStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// PatchDeployment returns a handler function for the patchDeployment tool
func PatchDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		patch, exists := args["patch"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: patch")
		}
		patchStr, ok := patch.(string)
		if !ok || patchStr == "" {
			return nil, fmt.Errorf("patch must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		patchType := "strategic"
		if pt, exists := args["patchType"]; exists {
			if ptStr, ok := pt.(string); ok && ptStr != "" {
				patchType = ptStr
			}
		}

		// Convert patch type string to k8s patch type
		var k8sPatchType types.PatchType
		switch patchType {
		case "json":
			k8sPatchType = types.JSONPatchType
		case "merge":
			k8sPatchType = types.MergePatchType
		case "strategic":
			k8sPatchType = types.StrategicMergePatchType
		default:
			k8sPatchType = types.StrategicMergePatchType
		}

		deployment, err := client.PatchDeployment(ctx, nameStr, namespace, []byte(patchStr), k8sPatchType)
		if err != nil {
			return nil, fmt.Errorf("failed to patch deployment: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Deployment '%s' patched successfully", nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"patchType":  patchType,
			"generation": deployment.Generation,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetDeploymentYAML returns a handler function for the getDeploymentYAML tool
func GetDeploymentYAML(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		export := false
		if exp, exists := args["export"]; exists {
			if expBool, ok := exp.(bool); ok {
				export = expBool
			}
		}

		yamlData, err := client.GetDeploymentYAML(ctx, nameStr, namespace, export)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment YAML: %v", err)
		}

		response := map[string]interface{}{
			"deployment": nameStr,
			"namespace":  namespace,
			"export":     export,
			"yaml":       yamlData,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// SetDeploymentResources returns a handler function for the setDeploymentResources tool
func SetDeploymentResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		container, exists := args["container"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: container")
		}
		containerStr, ok := container.(string)
		if !ok || containerStr == "" {
			return nil, fmt.Errorf("container must be a non-empty string")
		}

		resources, exists := args["resources"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: resources")
		}
		resourcesStr, ok := resources.(string)
		if !ok || resourcesStr == "" {
			return nil, fmt.Errorf("resources must be a non-empty string")
		}

		// Parse resources JSON
		var resourceRequirements corev1.ResourceRequirements
		err := json.Unmarshal([]byte(resourcesStr), &resourceRequirements)
		if err != nil {
			return nil, fmt.Errorf("invalid resources JSON: %v", err)
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		deployment, err := client.SetDeploymentResources(ctx, nameStr, namespace, containerStr, resourceRequirements)
		if err != nil {
			return nil, fmt.Errorf("failed to set deployment resources: %v", err)
		}

		response := map[string]interface{}{
			"message":    fmt.Sprintf("Resources updated for container '%s' in deployment '%s'", containerStr, nameStr),
			"deployment": nameStr,
			"namespace":  namespace,
			"container":  containerStr,
			"generation": deployment.Generation,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetDeploymentMetrics returns a handler function for the getDeploymentMetrics tool
func GetDeploymentMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		name, exists := args["name"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: name")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr == "" {
			return nil, fmt.Errorf("name must be a non-empty string")
		}

		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		metrics, err := client.GetDeploymentMetrics(ctx, nameStr, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment metrics: %v", err)
		}

		jsonResponse, err := json.Marshal(metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ListAllDeployments returns a handler function for the listAllDeployments tool
func ListAllDeployments(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		labelSelector := ""
		if selector, exists := args["labelSelector"]; exists {
			if selectorStr, ok := selector.(string); ok {
				labelSelector = selectorStr
			}
		}

		includeSystem := false
		if include, exists := args["includeSystem"]; exists {
			if includeBool, ok := include.(bool); ok {
				includeSystem = includeBool
			}
		}

		deployments, err := client.ListAllDeployments(ctx, labelSelector, includeSystem)
		if err != nil {
			return nil, fmt.Errorf("failed to list all deployments: %v", err)
		}

		jsonResponse, err := json.Marshal(deployments)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ScaleAllDeployments returns a handler function for the scaleAllDeployments tool
func ScaleAllDeployments(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		replicas, exists := args["replicas"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: replicas")
		}
		var replicasInt32 int32
		switch v := replicas.(type) {
		case float64:
			replicasInt32 = int32(v)
		case int:
			replicasInt32 = int32(v)
		case int32:
			replicasInt32 = v
		default:
			return nil, fmt.Errorf("replicas must be a number")
		}

		labelSelector := ""
		if selector, exists := args["labelSelector"]; exists {
			if selectorStr, ok := selector.(string); ok {
				labelSelector = selectorStr
			}
		}

		dryRun := false
		if dry, exists := args["dryRun"]; exists {
			if dryBool, ok := dry.(bool); ok {
				dryRun = dryBool
			}
		}

		result, err := client.ScaleAllDeployments(ctx, namespaceStr, replicasInt32, labelSelector, dryRun)
		if err != nil {
			return nil, fmt.Errorf("failed to scale all deployments: %v", err)
		}

		jsonResponse, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNamespaceResourceUsage returns a handler function for the getNamespaceResourceUsage tool
func GetNamespaceResourceUsage(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		namespace, exists := args["namespace"]
		if !exists {
			return nil, fmt.Errorf("missing required argument: namespace")
		}
		namespaceStr, ok := namespace.(string)
		if !ok || namespaceStr == "" {
			return nil, fmt.Errorf("namespace must be a non-empty string")
		}

		includeMetrics := false
		if metrics, exists := args["includeMetrics"]; exists {
			if metricsBool, ok := metrics.(bool); ok {
				includeMetrics = metricsBool
			}
		}

		usage, err := client.GetNamespaceResourceUsage(ctx, namespaceStr, includeMetrics)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace resource usage: %v", err)
		}

		jsonResponse, err := json.Marshal(usage)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetClusterOverview returns a handler function for the getClusterOverview tool
func GetClusterOverview(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("kubernetes client not available")
		}

		args := getArguments(request)

		includeMetrics := false
		if metrics, exists := args["includeMetrics"]; exists {
			if metricsBool, ok := metrics.(bool); ok {
				includeMetrics = metricsBool
			}
		}

		overview, err := client.GetClusterOverview(ctx, includeMetrics)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster overview: %v", err)
		}

		jsonResponse, err := json.Marshal(overview)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
// ========== SERVICE HANDLERS ==========

// ListServices returns a handler function for the listServices tool
func ListServices(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)
        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        var services []map[string]interface{}
        var err error

        if labelSelector, exists := args["labelSelector"]; exists {
            if selectorStr, ok := labelSelector.(string); ok && selectorStr != "" {
                services, err = client.ListServicesWithSelector(ctx, namespace, selectorStr)
            } else {
                services, err = client.ListServices(ctx, namespace)
            }
        } else {
            services, err = client.ListServices(ctx, namespace)
        }

        if err != nil {
            return nil, fmt.Errorf("failed to list services: %v", err)
        }

        response := map[string]interface{}{
            "namespace": namespace,
            "services":  services,
            "count":     len(services),
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// GetService returns a handler function for the getService tool
func GetService(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)
        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        service, err := client.GetService(ctx, nameStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to get service: %v", err)
        }

        jsonResponse, err := json.Marshal(service)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// CreateService returns a handler function for the createService tool
func CreateService(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        manifest, exists := args["manifest"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: manifest")
        }
        manifestStr, ok := manifest.(string)
        if !ok || manifestStr == "" {
            return nil, fmt.Errorf("manifest must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        service, err := client.CreateService(ctx, manifestStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to create service: %v", err)
        }

        response := map[string]interface{}{
            "message": fmt.Sprintf("Service '%s' created successfully in namespace '%s'", service.Name, service.Namespace),
            "service": map[string]interface{}{
                "name":              service.Name,
                "namespace":         service.Namespace,
                "uid":               service.UID,
                "type":              string(service.Spec.Type),
                "clusterIP":         service.Spec.ClusterIP,
                "ports":             service.Spec.Ports,
                "selector":          service.Spec.Selector,
                "creationTimestamp": service.CreationTimestamp.Time.Format(time.RFC3339),
            },
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// UpdateService returns a handler function for the updateService tool
func UpdateService(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        manifest, exists := args["manifest"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: manifest")
        }
        manifestStr, ok := manifest.(string)
        if !ok || manifestStr == "" {
            return nil, fmt.Errorf("manifest must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        service, err := client.UpdateService(ctx, nameStr, manifestStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to update service: %v", err)
        }

        response := map[string]interface{}{
            "message": fmt.Sprintf("Service '%s' updated successfully", service.Name),
            "service": map[string]interface{}{
                "name":             service.Name,
                "namespace":        service.Namespace,
                "resourceVersion":  service.ResourceVersion,
                "type":             string(service.Spec.Type),
                "clusterIP":        service.Spec.ClusterIP,
                "ports":            service.Spec.Ports,
                "selector":         service.Spec.Selector,
            },
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// DeleteService returns a handler function for the deleteService tool
func DeleteService(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        err := client.DeleteService(ctx, nameStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to delete service: %v", err)
        }

        response := map[string]interface{}{
            "message":     fmt.Sprintf("Service '%s' deleted successfully", nameStr),
            "serviceName": nameStr,
            "namespace":   namespace,
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// GetServiceEndpoints returns a handler function for the getServiceEndpoints tool
func GetServiceEndpoints(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        endpoints, err := client.GetServiceEndpoints(ctx, nameStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to get service endpoints: %v", err)
        }

        jsonResponse, err := json.Marshal(endpoints)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// TestServiceConnectivity returns a handler function for the testServiceConnectivity tool
func TestServiceConnectivity(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        var port int32 = 0
        if portArg, exists := args["port"]; exists {
            switch v := portArg.(type) {
            case float64:
                port = int32(v)
            case int:
                port = int32(v)
            case int32:
                port = v
            }
        }

        protocol := "TCP"
        if protocolArg, exists := args["protocol"]; exists {
            if protocolStr, ok := protocolArg.(string); ok && protocolStr != "" {
                protocol = protocolStr
            }
        }

        connectivity, err := client.TestServiceConnectivity(ctx, nameStr, namespace, port, protocol)
        if err != nil {
            return nil, fmt.Errorf("failed to test service connectivity: %v", err)
        }

        jsonResponse, err := json.Marshal(connectivity)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// ========== EXTENDED SERVICE HANDLERS ==========

// GetServiceEvents returns a handler function for the getServiceEvents tool
func GetServiceEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        limit := int64(50)
        if limitArg, exists := args["limit"]; exists {
            switch v := limitArg.(type) {
            case float64:
                limit = int64(v)
            case int:
                limit = int64(v)
            case int64:
                limit = v
            }
        }

        events, err := client.GetServiceEvents(ctx, nameStr, namespace, limit)
        if err != nil {
            return nil, fmt.Errorf("failed to get service events: %v", err)
        }

        response := map[string]interface{}{
            "serviceName": nameStr,
            "namespace":   namespace,
            "events":      events,
            "count":       len(events),
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// GetServiceYAML returns a handler function for the getServiceYAML tool
func GetServiceYAML(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        export := false
        if exp, exists := args["export"]; exists {
            if expBool, ok := exp.(bool); ok {
                export = expBool
            }
        }

        yamlData, err := client.GetServiceYAML(ctx, nameStr, namespace, export)
        if err != nil {
            return nil, fmt.Errorf("failed to get service YAML: %v", err)
        }

        response := map[string]interface{}{
            "serviceName": nameStr,
            "namespace":   namespace,
            "export":      export,
            "yaml":        yamlData,
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// ExposeDeployment returns a handler function for the exposeDeployment tool
func ExposeDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        deployment, exists := args["deployment"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: deployment")
        }
        deploymentStr, ok := deployment.(string)
        if !ok || deploymentStr == "" {
            return nil, fmt.Errorf("deployment must be a non-empty string")
        }

        port, exists := args["port"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: port")
        }
        var portInt32 int32
        switch v := port.(type) {
        case float64:
            portInt32 = int32(v)
        case int:
            portInt32 = int32(v)
        case int32:
            portInt32 = v
        default:
            return nil, fmt.Errorf("port must be a number")
        }

        serviceName := deploymentStr
        if sn, exists := args["serviceName"]; exists {
            if snStr, ok := sn.(string); ok && snStr != "" {
                serviceName = snStr
            }
        }

        var targetPort int32 = portInt32
        if tp, exists := args["targetPort"]; exists {
            switch v := tp.(type) {
            case float64:
                targetPort = int32(v)
            case int:
                targetPort = int32(v)
            case int32:
                targetPort = v
            }
        }

        serviceType := "ClusterIP"
        if st, exists := args["serviceType"]; exists {
            if stStr, ok := st.(string); ok && stStr != "" {
                serviceType = stStr
            }
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        service, err := client.ExposeDeployment(ctx, deploymentStr, serviceName, namespace, portInt32, targetPort, serviceType)
        if err != nil {
            return nil, fmt.Errorf("failed to expose deployment: %v", err)
        }

        response := map[string]interface{}{
            "message": fmt.Sprintf("Deployment '%s' exposed as service '%s'", deploymentStr, serviceName),
            "service": map[string]interface{}{
                "name":       service.Name,
                "namespace":  service.Namespace,
                "type":       string(service.Spec.Type),
                "clusterIP":  service.Spec.ClusterIP,
                "ports":      service.Spec.Ports,
                "selector":   service.Spec.Selector,
            },
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}
// PatchService returns a handler function for the patchService tool
func PatchService(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        patch, exists := args["patch"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: patch")
        }
        patchStr, ok := patch.(string)
        if !ok || patchStr == "" {
            return nil, fmt.Errorf("patch must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        patchType := "strategic"
        if pt, exists := args["patchType"]; exists {
            if ptStr, ok := pt.(string); ok && ptStr != "" {
                patchType = ptStr
            }
        }

        // Convert patch type string to k8s patch type
        var k8sPatchType types.PatchType
        switch patchType {
        case "json":
            k8sPatchType = types.JSONPatchType
        case "merge":
            k8sPatchType = types.MergePatchType
        case "strategic":
            k8sPatchType = types.StrategicMergePatchType
        default:
            k8sPatchType = types.StrategicMergePatchType
        }

        service, err := client.PatchService(ctx, nameStr, namespace, []byte(patchStr), k8sPatchType)
        if err != nil {
            return nil, fmt.Errorf("failed to patch service: %v", err)
        }

        response := map[string]interface{}{
            "message":     fmt.Sprintf("Service '%s' patched successfully", nameStr),
            "serviceName": nameStr,
            "namespace":   namespace,
            "patchType":   patchType,
            "service": map[string]interface{}{
                "name":            service.Name,
                "resourceVersion": service.ResourceVersion,
                "type":            string(service.Spec.Type),
                "clusterIP":       service.Spec.ClusterIP,
            },
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// ListAllServices returns a handler function for the listAllServices tool
func ListAllServices(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        labelSelector := ""
        if selector, exists := args["labelSelector"]; exists {
            if selectorStr, ok := selector.(string); ok {
                labelSelector = selectorStr
            }
        }

        includeSystem := false
        if include, exists := args["includeSystem"]; exists {
            if includeBool, ok := include.(bool); ok {
                includeSystem = includeBool
            }
        }

        services, err := client.ListAllServices(ctx, labelSelector, includeSystem)
        if err != nil {
            return nil, fmt.Errorf("failed to list all services: %v", err)
        }

        jsonResponse, err := json.Marshal(services)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// GetServiceMetrics returns a handler function for the getServiceMetrics tool
func GetServiceMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        metrics, err := client.GetServiceMetrics(ctx, nameStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to get service metrics: %v", err)
        }

        jsonResponse, err := json.Marshal(metrics)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// GetServiceTopology returns a handler function for the getServiceTopology tool
func GetServiceTopology(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        name, exists := args["name"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: name")
        }
        nameStr, ok := name.(string)
        if !ok || nameStr == "" {
            return nil, fmt.Errorf("name must be a non-empty string")
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        topology, err := client.GetServiceTopology(ctx, nameStr, namespace)
        if err != nil {
            return nil, fmt.Errorf("failed to get service topology: %v", err)
        }

        jsonResponse, err := json.Marshal(topology)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}

// CreateServiceFromPods returns a handler function for the createServiceFromPods tool
func CreateServiceFromPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        if client == nil {
            return nil, fmt.Errorf("kubernetes client not available")
        }

        args := getArguments(request)

        serviceName, exists := args["serviceName"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: serviceName")
        }
        serviceNameStr, ok := serviceName.(string)
        if !ok || serviceNameStr == "" {
            return nil, fmt.Errorf("serviceName must be a non-empty string")
        }

        labelSelector, exists := args["labelSelector"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: labelSelector")
        }
        labelSelectorStr, ok := labelSelector.(string)
        if !ok || labelSelectorStr == "" {
            return nil, fmt.Errorf("labelSelector must be a non-empty string")
        }

        port, exists := args["port"]
        if !exists {
            return nil, fmt.Errorf("missing required argument: port")
        }
        var portInt32 int32
        switch v := port.(type) {
        case float64:
            portInt32 = int32(v)
        case int:
            portInt32 = int32(v)
        case int32:
            portInt32 = v
        default:
            return nil, fmt.Errorf("port must be a number")
        }

        var targetPort int32 = portInt32
        if tp, exists := args["targetPort"]; exists {
            switch v := tp.(type) {
            case float64:
                targetPort = int32(v)
            case int:
                targetPort = int32(v)
            case int32:
                targetPort = v
            }
        }

        serviceType := "ClusterIP"
        if st, exists := args["serviceType"]; exists {
            if stStr, ok := st.(string); ok && stStr != "" {
                serviceType = stStr
            }
        }

        namespace := "default"
        if ns, exists := args["namespace"]; exists {
            if nsStr, ok := ns.(string); ok && nsStr != "" {
                namespace = nsStr
            }
        }

        service, err := client.CreateServiceFromPods(ctx, serviceNameStr, namespace, labelSelectorStr, portInt32, targetPort, serviceType)
        if err != nil {
            return nil, fmt.Errorf("failed to create service from pods: %v", err)
        }

        response := map[string]interface{}{
            "message": fmt.Sprintf("Service '%s' created successfully from pod selector '%s'", serviceNameStr, labelSelectorStr),
            "service": map[string]interface{}{
                "name":          service.Name,
                "namespace":     service.Namespace,
                "type":          string(service.Spec.Type),
                "clusterIP":     service.Spec.ClusterIP,
                "ports":         service.Spec.Ports,
                "selector":      service.Spec.Selector,
                "labelSelector": labelSelectorStr,
            },
        }

        jsonResponse, err := json.Marshal(response)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize response: %v", err)
        }

        return mcp.NewToolResultText(string(jsonResponse)), nil
    }
}