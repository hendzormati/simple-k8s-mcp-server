package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListPods returns a handler function for the listPods tool
func ListPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments from the request (no type assertion needed)
		args := request.Params.Arguments
		if args == nil {
			args = make(map[string]interface{})
		}

		// Get namespace argument (default to "default")
		namespace := "default"
		if ns, exists := args["namespace"]; exists {
			if nsStr, ok := ns.(string); ok && nsStr != "" {
				namespace = nsStr
			}
		}

		// Get pods from Kubernetes
		pods, err := client.GetPodsInNamespace(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods: %v", err)
		}

		// Prepare response
		response := map[string]interface{}{
			"namespace": namespace,
			"pods":      pods,
			"count":     len(pods),
		}

		// Convert to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %v", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ========== NAMESPACE HANDLERS ==========

// ListNamespaces returns a handler function for the listNamespaces tool
func ListNamespaces(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if Kubernetes client is available
		if client == nil {
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
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
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := request.Params.Arguments
		if args == nil {
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
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := request.Params.Arguments
		if args == nil {
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

		// Get optional labels
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsMap, ok := labelsArg.(map[string]interface{}); ok {
				labels = make(map[string]string)
				for k, v := range labelsMap {
					if vStr, ok := v.(string); ok {
						labels[k] = vStr
					}
				}
			}
		}

		// Get optional annotations
		var annotations map[string]string
		if annotationsArg, exists := args["annotations"]; exists {
			if annotationsMap, ok := annotationsArg.(map[string]interface{}); ok {
				annotations = make(map[string]string)
				for k, v := range annotationsMap {
					if vStr, ok := v.(string); ok {
						annotations[k] = vStr
					}
				}
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
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := request.Params.Arguments
		if args == nil {
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

		// Get optional labels
		var labels map[string]string
		if labelsArg, exists := args["labels"]; exists {
			if labelsMap, ok := labelsArg.(map[string]interface{}); ok {
				labels = make(map[string]string)
				for k, v := range labelsMap {
					if vStr, ok := v.(string); ok {
						labels[k] = vStr
					}
				}
			}
		}

		// Get optional annotations
		var annotations map[string]string
		if annotationsArg, exists := args["annotations"]; exists {
			if annotationsMap, ok := annotationsArg.(map[string]interface{}); ok {
				annotations = make(map[string]string)
				for k, v := range annotationsMap {
					if vStr, ok := v.(string); ok {
						annotations[k] = vStr
					}
				}
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
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := request.Params.Arguments
		if args == nil {
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
			return nil, fmt.Errorf("Kubernetes client not available - please configure a Kubernetes cluster")
		}

		// Extract arguments
		args := request.Params.Arguments
		if args == nil {
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
