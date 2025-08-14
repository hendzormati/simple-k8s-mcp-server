package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
	"github.com/mark3labs/mcp-go/mcp"
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

		// Delete namespace
		err := client.DeleteNamespace(ctx, nameStr)
		if err != nil {
			return nil, fmt.Errorf("failed to delete namespace: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"message":   fmt.Sprintf("Namespace '%s' has been deleted successfully", nameStr),
			"namespace": nameStr,
			"status":    "deleted",
		}

		// Convert to JSON
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

// ========== POD HANDLERS ==========

// ListPods returns a handler function for the listPods tool
func ListPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client == nil {
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
			return nil, fmt.Errorf("Kubernetes client not available")
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
