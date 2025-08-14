package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ListPodsTool creates a tool for listing pods in a namespace
func ListPodsTool() mcp.Tool {
	return mcp.NewTool(
		"listPods",
		mcp.WithDescription("List all pods in a Kubernetes namespace"),
		mcp.WithString("namespace", mcp.Description("The namespace to list pods from (default: 'default')")),
	)
}

// ========== NAMESPACE TOOLS ==========

// ListNamespacesTool creates a tool for listing all namespaces
func ListNamespacesTool() mcp.Tool {
	return mcp.NewTool(
		"listNamespaces",
		mcp.WithDescription("List all namespaces in the Kubernetes cluster"),
	)
}

// GetNamespaceTool creates a tool for getting detailed information about a specific namespace
func GetNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespace",
		mcp.WithDescription("Get detailed information about a specific namespace"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to retrieve")),
	)
}

// CreateNamespaceTool creates a tool for creating a new namespace
func CreateNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"createNamespace",
		mcp.WithDescription("Create a new namespace with optional labels and annotations"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to create")),
		mcp.WithString("labels", mcp.Description("Optional labels for the namespace in JSON format (e.g., '{\"env\":\"dev\",\"team\":\"backend\"}')")),
		mcp.WithString("annotations", mcp.Description("Optional annotations for the namespace in JSON format (e.g., '{\"description\":\"Development namespace\"}')")),
	)
}

// UpdateNamespaceTool creates a tool for updating namespace labels and annotations
func UpdateNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"updateNamespace",
		mcp.WithDescription("Update labels and annotations of an existing namespace"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to update")),
		mcp.WithString("labels", mcp.Description("Labels to set on the namespace in JSON format (e.g., '{\"env\":\"prod\",\"version\":\"v2\"}')")),
		mcp.WithString("annotations", mcp.Description("Annotations to set on the namespace in JSON format (e.g., '{\"owner\":\"team-alpha\"}')")),
	)
}

// DeleteNamespaceTool creates a tool for deleting a namespace
func DeleteNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"deleteNamespace",
		mcp.WithDescription("Delete a namespace (WARNING: This will delete all resources in the namespace)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to delete")),
	)
}

// GetNamespaceResourceQuotaTool creates a tool for getting resource quotas in a namespace
func GetNamespaceResourceQuotaTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceResourceQuota",
		mcp.WithDescription("Get resource quotas for a specific namespace"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to get resource quotas from")),
	)
}

// Add these tools to your k8s.go file in the tools package

// GetNamespaceEventsTool creates a tool for getting events in a namespace
func GetNamespaceEventsTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceEvents",
		mcp.WithDescription("Get all events in a specific namespace to diagnose issues"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to get events from")),
	)
}

// GetNamespaceAllResourcesTool creates a tool for getting all resources in a namespace
func GetNamespaceAllResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceAllResources",
		mcp.WithDescription("Get all resources in a namespace to identify what might be blocking deletion"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to get all resources from")),
	)
}

// ForceDeleteNamespaceTool creates a tool for force deleting a namespace
func ForceDeleteNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"forceDeleteNamespace",
		mcp.WithDescription("Force delete a namespace by removing finalizers (use with caution)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to force delete")),
	)
}
