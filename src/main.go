package main

import (
	"fmt"
	"log"

	"github.com/hendzormati/simple-k8s-mcp-server/handlers"
	"github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
	"github.com/hendzormati/simple-k8s-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	fmt.Println("Starting simple K8s MCP server...")

	// Initialize Kubernetes client (with graceful error handling)
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Printf("Warning: Failed to create K8s client: %v", err)
		log.Println("Server will start but K8s features won't work")
		log.Println("To fix this: Set up a Kubernetes cluster or configure kubeconfig")
	} else {
		// Test connection
		if err := k8sClient.TestConnection(); err != nil {
			log.Printf("Warning: Cannot connect to K8s cluster: %v", err)
			log.Println("Server will start but K8s features won't work")
		} else {
			fmt.Println("✅ Successfully connected to Kubernetes cluster!")
		}
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"Simple K8s MCP Server",
		"1.0.0",
	)

	// Register Pod tools
	mcpServer.AddTool(tools.ListPodsTool(), handlers.ListPods(k8sClient))

	// Register Namespace tools
	mcpServer.AddTool(tools.ListNamespacesTool(), handlers.ListNamespaces(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceTool(), handlers.GetNamespace(k8sClient))
	mcpServer.AddTool(tools.CreateNamespaceTool(), handlers.CreateNamespace(k8sClient))
	mcpServer.AddTool(tools.UpdateNamespaceTool(), handlers.UpdateNamespace(k8sClient))
	mcpServer.AddTool(tools.DeleteNamespaceTool(), handlers.DeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceResourceQuotaTool(), handlers.GetNamespaceResourceQuota(k8sClient))

	fmt.Println("MCP Server initialized with tools:")
	fmt.Println("  Pod Tools:")
	fmt.Println("    - listPods: List pods in a namespace")
	fmt.Println("  Namespace Tools:")
	fmt.Println("    - listNamespaces: List all namespaces")
	fmt.Println("    - getNamespace: Get detailed namespace information")
	fmt.Println("    - createNamespace: Create a new namespace")
	fmt.Println("    - updateNamespace: Update namespace labels/annotations")
	fmt.Println("    - deleteNamespace: Delete a namespace")
	fmt.Println("    - getNamespaceResourceQuota: Get resource quotas for a namespace")
	fmt.Println()
	fmt.Println("Server is ready and listening for MCP protocol messages...")

	// Start server in stdio mode (for MCP protocol)
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}
