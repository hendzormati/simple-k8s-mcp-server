package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hendzormati/simple-k8s-mcp-server/handlers"
	"github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
	"github.com/hendzormati/simple-k8s-mcp-server/tools"
	"github.com/mark3labs/mcp-go/server"
)

// getEnvOrDefault returns the value of the environment variable or the default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	fmt.Println("Starting simple K8s MCP server...")

	// Parse command line flags
	var mode string
	var port string
	var host string

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8080"), "Server port")
	flag.StringVar(&host, "host", getEnvOrDefault("SERVER_HOST", "localhost"), "Server host address")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio' or 'sse'")
	flag.Parse()

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
		server.WithResourceCapabilities(true, true), // Enable resource listing and subscription capabilities
	)

	// Pod tools
	mcpServer.AddTool(tools.ListPodsTool(), handlers.ListPods(k8sClient))
	mcpServer.AddTool(tools.GetPodTool(), handlers.GetPod(k8sClient))
	mcpServer.AddTool(tools.GetPodLogsTool(), handlers.GetPodLogs(k8sClient))
	mcpServer.AddTool(tools.GetPodMetricsTool(), handlers.GetPodMetrics(k8sClient))
	mcpServer.AddTool(tools.DescribePodTool(), handlers.DescribePod(k8sClient))
	mcpServer.AddTool(tools.DeletePodTool(), handlers.DeletePod(k8sClient))
	mcpServer.AddTool(tools.GetPodEventsTool(), handlers.GetPodEvents(k8sClient))
	mcpServer.AddTool(tools.RestartPodTool(), handlers.RestartPod(k8sClient))
	mcpServer.AddTool(tools.CreatePodTool(), handlers.CreatePod(k8sClient))
	mcpServer.AddTool(tools.UpdatePodTool(), handlers.UpdatePod(k8sClient))

	// Namespace tools
	mcpServer.AddTool(tools.ListNamespacesTool(), handlers.ListNamespaces(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceTool(), handlers.GetNamespace(k8sClient))
	mcpServer.AddTool(tools.CreateNamespaceTool(), handlers.CreateNamespace(k8sClient))
	mcpServer.AddTool(tools.UpdateNamespaceTool(), handlers.UpdateNamespace(k8sClient))
	mcpServer.AddTool(tools.DeleteNamespaceTool(), handlers.DeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.ForceDeleteNamespaceTool(), handlers.ForceDeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.SmartDeleteNamespaceTool(), handlers.SmartDeleteNamespace(k8sClient)) // Add this line
	mcpServer.AddTool(tools.GetNamespaceResourceQuotaTool(), handlers.GetNamespaceResourceQuota(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceEventsTool(), handlers.GetNamespaceEvents(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceAllResourcesTool(), handlers.GetNamespaceAllResources(k8sClient))
	mcpServer.AddTool(tools.ForceDeleteNamespaceTool(), handlers.ForceDeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceYAMLTool(), handlers.GetNamespaceYAML(k8sClient))
	mcpServer.AddTool(tools.SetNamespaceResourceQuotaTool(), handlers.SetNamespaceResourceQuota(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceLimitRangesTool(), handlers.GetNamespaceLimitRanges(k8sClient))
	mcpServer.AddTool(tools.SetNamespaceLimitRangeTool(), handlers.SetNamespaceLimitRange(k8sClient))

	// Deployment tools
	mcpServer.AddTool(tools.ListDeploymentsTool(), handlers.ListDeployments(k8sClient))
	mcpServer.AddTool(tools.GetDeploymentTool(), handlers.GetDeployment(k8sClient))
	mcpServer.AddTool(tools.CreateDeploymentTool(), handlers.CreateDeployment(k8sClient))
	mcpServer.AddTool(tools.UpdateDeploymentTool(), handlers.UpdateDeployment(k8sClient))
	mcpServer.AddTool(tools.DeleteDeploymentTool(), handlers.DeleteDeployment(k8sClient))
	mcpServer.AddTool(tools.ScaleDeploymentTool(), handlers.ScaleDeployment(k8sClient))
	mcpServer.AddTool(tools.RolloutStatusTool(), handlers.RolloutStatus(k8sClient))
	mcpServer.AddTool(tools.RolloutHistoryTool(), handlers.RolloutHistory(k8sClient))
	mcpServer.AddTool(tools.RolloutUndoTool(), handlers.RolloutUndo(k8sClient))
	mcpServer.AddTool(tools.PauseDeploymentTool(), handlers.PauseDeployment(k8sClient))
	mcpServer.AddTool(tools.ResumeDeploymentTool(), handlers.ResumeDeployment(k8sClient))

	fmt.Println("MCP Server initialized with tools:")
	fmt.Println("  Pod Tools:")
	fmt.Println("    - listPods: List pods in a namespace with filtering")
	fmt.Println("    - getPod: Get detailed information about a specific pod")
	fmt.Println("    - getPodLogs: Get logs from a pod")
	fmt.Println("    - getPodMetrics: Get CPU and memory metrics for a pod")
	fmt.Println("    - describePod: Get comprehensive pod description")
	fmt.Println("    - deletePod: Delete a specific pod")
	fmt.Println("    - getPodEvents: Get events related to a pod")
	fmt.Println("    - restartPod: Restart a pod by deleting it")
	fmt.Println("    - createPod: Create a new pod from JSON manifest")
	fmt.Println("    - updatePod: Update pod labels and annotations")
	fmt.Println("  Namespace Tools:")
	fmt.Println("    - listNamespaces: List all namespaces")
	fmt.Println("    - getNamespace: Get detailed namespace information")
	fmt.Println("    - createNamespace: Create a new namespace")
	fmt.Println("    - updateNamespace: Update namespace labels/annotations")
	fmt.Println("    - deleteNamespace: Delete a namespace")
	fmt.Println("    - forceDeleteNamespace: Force delete a stuck namespace")
	fmt.Println("    - smartDeleteNamespace: Intelligently delete using best strategy")
	fmt.Println("    - getNamespaceResourceQuota: Get resource quotas for a namespace")
	fmt.Println("    - getNamespaceYAML: Get namespace YAML definition")
	fmt.Println("    - setNamespaceResourceQuota: Set resource quota in namespace")
	fmt.Println("    - getNamespaceLimitRanges: Get limit ranges in namespace")
	fmt.Println("    - setNamespaceLimitRange: Set limit range in namespace")
	fmt.Println("  📦 Deployment Tools:")
	fmt.Println("    - listDeployments: List deployments in namespace")
	fmt.Println("    - getDeployment: Get deployment details")
	fmt.Println("    - createDeployment: Create new deployments")
	fmt.Println("    - updateDeployment: Update deployment specs")
	fmt.Println("    - deleteDeployment: Delete deployments")
	fmt.Println("    - scaleDeployment: Scale replicas up/down")
	fmt.Println("    - rolloutStatus: Check rollout status")
	fmt.Println("    - rolloutHistory: Get rollout history")
	fmt.Println("    - rolloutUndo: Rollback to previous version")
	fmt.Println("    - pauseDeployment: Pause deployments")
	fmt.Println("    - resumeDeployment: Resume deployments")
	fmt.Println()

	// Start server based on mode
	switch mode {
	case "stdio":
		fmt.Println("Starting server in stdio mode...")
		fmt.Println("Server is ready and listening for MCP protocol messages...")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Failed to start stdio server: %v", err)
		}
	case "sse":
		address := fmt.Sprintf("%s:%s", host, port)
		fmt.Printf("Starting server in SSE mode on %s...\n", address)

		// Create SSE server
		sse := server.NewSSEServer(mcpServer)

		// Start server in a goroutine
		go func() {
			if err := sse.Start(address); err != nil {
				log.Printf("Failed to start SSE server: %v", err)
				os.Exit(1)
			}
		}()

		fmt.Printf("SSE server started on %s\n", address)
		fmt.Printf("Connect to: http://%s/sse\n", address)
		fmt.Printf("Message endpoint: http://%s/sse/message?sessionId=<session-id>\n", address)
		fmt.Println("Press Ctrl+C to stop the server...")

		// Set up signal handling for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Block until we receive a signal
		<-c
		fmt.Println("\nShutting down server...")

	default:
		fmt.Printf("Unknown server mode: %s. Use 'stdio' or 'sse'.\n", mode)
		return
	}
}
