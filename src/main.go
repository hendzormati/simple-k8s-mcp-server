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
	fmt.Println("🚀 Starting Simple K8s MCP Server...")

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
		log.Printf("⚠️  Warning: Failed to create K8s client: %v", err)
		log.Println("📋 Server will start but K8s features won't work")
		log.Println("💡 To fix this: Set up a Kubernetes cluster or configure kubeconfig")
	} else {
		// Test connection
		if err := k8sClient.TestConnection(); err != nil {
			log.Printf("⚠️  Warning: Cannot connect to K8s cluster: %v", err)
			log.Println("📋 Server will start but K8s features won't work")
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

	// Register all tools
	registerAllTools(mcpServer, k8sClient)

	// Print available tools in organized format
	printToolsOverview()

	fmt.Println()

	// Start server based on mode
	switch mode {
	case "stdio":
		fmt.Println("🎯 Starting server in stdio mode...")
		fmt.Println("📡 Server is ready and listening for MCP protocol messages...")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("❌ Failed to start stdio server: %v", err)
		}
	case "sse":
		address := fmt.Sprintf("%s:%s", host, port)
		fmt.Printf("🌐 Starting server in SSE mode on %s...\n", address)

		// Create SSE server
		sse := server.NewSSEServer(mcpServer)

		// Start server in a goroutine
		go func() {
			if err := sse.Start(address); err != nil {
				log.Printf("❌ Failed to start SSE server: %v", err)
				os.Exit(1)
			}
		}()

		fmt.Printf("✅ SSE server started on %s\n", address)
		fmt.Printf("🔗 Connect to: http://%s/sse\n", address)
		fmt.Printf("💬 Message endpoint: http://%s/sse/message?sessionId=<session-id>\n", address)
		fmt.Println("⏹️  Press Ctrl+C to stop the server...")

		// Set up signal handling for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Block until we receive a signal
		<-c
		fmt.Println("\n🛑 Shutting down server...")

	default:
		fmt.Printf("❌ Unknown server mode: %s. Use 'stdio' or 'sse'.\n", mode)
		return
	}
}

func registerAllTools(mcpServer *server.MCPServer, k8sClient *k8s.Client) {
	// Core Pod tools
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

	// Extended Pod tools
	mcpServer.AddTool(tools.GetPodResourceUsageTool(), handlers.GetPodResourceUsage(k8sClient))
	mcpServer.AddTool(tools.GetPodsHealthStatusTool(), handlers.GetPodsHealthStatus(k8sClient))

	// Core Namespace tools
	mcpServer.AddTool(tools.ListNamespacesTool(), handlers.ListNamespaces(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceTool(), handlers.GetNamespace(k8sClient))
	mcpServer.AddTool(tools.CreateNamespaceTool(), handlers.CreateNamespace(k8sClient))
	mcpServer.AddTool(tools.UpdateNamespaceTool(), handlers.UpdateNamespace(k8sClient))
	mcpServer.AddTool(tools.DeleteNamespaceTool(), handlers.DeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.ForceDeleteNamespaceTool(), handlers.ForceDeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.SmartDeleteNamespaceTool(), handlers.SmartDeleteNamespace(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceResourceQuotaTool(), handlers.GetNamespaceResourceQuota(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceEventsTool(), handlers.GetNamespaceEvents(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceAllResourcesTool(), handlers.GetNamespaceAllResources(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceYAMLTool(), handlers.GetNamespaceYAML(k8sClient))
	mcpServer.AddTool(tools.SetNamespaceResourceQuotaTool(), handlers.SetNamespaceResourceQuota(k8sClient))
	mcpServer.AddTool(tools.GetNamespaceLimitRangesTool(), handlers.GetNamespaceLimitRanges(k8sClient))
	mcpServer.AddTool(tools.SetNamespaceLimitRangeTool(), handlers.SetNamespaceLimitRange(k8sClient))

	// Extended Namespace tools
	mcpServer.AddTool(tools.GetNamespaceResourceUsageTool(), handlers.GetNamespaceResourceUsage(k8sClient))
	mcpServer.AddTool(tools.GetClusterOverviewTool(), handlers.GetClusterOverview(k8sClient))

	// Core Deployment tools
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

	// Extended Deployment tools
	mcpServer.AddTool(tools.GetDeploymentEventsTool(), handlers.GetDeploymentEvents(k8sClient))
	mcpServer.AddTool(tools.GetDeploymentLogsTool(), handlers.GetDeploymentLogs(k8sClient))
	mcpServer.AddTool(tools.RestartDeploymentTool(), handlers.RestartDeployment(k8sClient))
	mcpServer.AddTool(tools.WaitForDeploymentTool(), handlers.WaitForDeployment(k8sClient))
	mcpServer.AddTool(tools.SetDeploymentImageTool(), handlers.SetDeploymentImage(k8sClient))
	mcpServer.AddTool(tools.SetDeploymentEnvTool(), handlers.SetDeploymentEnv(k8sClient))
	mcpServer.AddTool(tools.PatchDeploymentTool(), handlers.PatchDeployment(k8sClient))
	mcpServer.AddTool(tools.GetDeploymentYAMLTool(), handlers.GetDeploymentYAML(k8sClient))
	mcpServer.AddTool(tools.SetDeploymentResourcesTool(), handlers.SetDeploymentResources(k8sClient))
	mcpServer.AddTool(tools.GetDeploymentMetricsTool(), handlers.GetDeploymentMetrics(k8sClient))
	mcpServer.AddTool(tools.ListAllDeploymentsTool(), handlers.ListAllDeployments(k8sClient))
	mcpServer.AddTool(tools.ScaleAllDeploymentsTool(), handlers.ScaleAllDeployments(k8sClient))
}

func printToolsOverview() {
	fmt.Println("🔧 ═══════════════════════════════════════════════════════════════")
	fmt.Println("📋 AVAILABLE KUBERNETES MCP TOOLS")
	fmt.Println("🔧 ═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Pod Management Section
	fmt.Println("🔵 POD MANAGEMENT")
	fmt.Println("  📊 Core Operations:")
	fmt.Println("    • listPods           - List pods in namespace with filtering")
	fmt.Println("    • getPod             - Get detailed pod information")
	fmt.Println("    • createPod          - Create new pod from manifest")
	fmt.Println("    • updatePod          - Update pod labels/annotations")
	fmt.Println("    • deletePod          - Delete specific pod")
	fmt.Println("    • restartPod         - Restart pod by deletion")
	fmt.Println()
	fmt.Println("  🔍 Monitoring & Debugging:")
	fmt.Println("    • describePod        - Comprehensive pod description")
	fmt.Println("    • getPodLogs         - Get container logs")
	fmt.Println("    • getPodEvents       - Get pod-related events")
	fmt.Println("    • getPodMetrics      - Get CPU/memory metrics")
	fmt.Println("    • getPodResourceUsage - Get resource usage details")
	fmt.Println()
	fmt.Println("  📈 Health & Status:")
	fmt.Println("    • getPodsHealthStatus - Health overview for multiple pods")
	fmt.Println()

	// Namespace Management Section
	fmt.Println("🟢 NAMESPACE MANAGEMENT")
	fmt.Println("  📊 Core Operations:")
	fmt.Println("    • listNamespaces         - List all namespaces")
	fmt.Println("    • getNamespace           - Get namespace details")
	fmt.Println("    • createNamespace        - Create new namespace")
	fmt.Println("    • updateNamespace        - Update labels/annotations")
	fmt.Println("    • deleteNamespace        - Standard namespace deletion")
	fmt.Println("    • forceDeleteNamespace   - Force delete stuck namespaces")
	fmt.Println("    • smartDeleteNamespace   - Auto-choose deletion strategy")
	fmt.Println()
	fmt.Println("  🎛️  Resource Management:")
	fmt.Println("    • getNamespaceResourceQuota  - Get resource quotas")
	fmt.Println("    • setNamespaceResourceQuota  - Set resource quotas")
	fmt.Println("    • getNamespaceLimitRanges    - Get limit ranges")
	fmt.Println("    • setNamespaceLimitRange     - Set limit ranges")
	fmt.Println("    • getNamespaceResourceUsage  - Resource usage summary")
	fmt.Println()
	fmt.Println("  🔍 Monitoring & Export:")
	fmt.Println("    • getNamespaceEvents        - Get namespace events")
	fmt.Println("    • getNamespaceAllResources  - List all resources")
	fmt.Println("    • getNamespaceYAML          - Export as YAML")
	fmt.Println()

	// Deployment Management Section
	fmt.Println("🟡 DEPLOYMENT MANAGEMENT")
	fmt.Println("  📊 Core Operations:")
	fmt.Println("    • listDeployments     - List deployments in namespace")
	fmt.Println("    • getDeployment       - Get deployment details")
	fmt.Println("    • createDeployment    - Create new deployment")
	fmt.Println("    • updateDeployment    - Update deployment specs")
	fmt.Println("    • deleteDeployment    - Delete deployment")
	fmt.Println()
	fmt.Println("  ⚡ Scaling & Rollouts:")
	fmt.Println("    • scaleDeployment     - Scale replicas up/down")
	fmt.Println("    • rolloutStatus       - Check rollout status")
	fmt.Println("    • rolloutHistory      - Get rollout history")
	fmt.Println("    • rolloutUndo         - Rollback to previous version")
	fmt.Println("    • pauseDeployment     - Pause deployment rollouts")
	fmt.Println("    • resumeDeployment    - Resume deployment rollouts")
	fmt.Println("    • restartDeployment   - Restart deployment")
	fmt.Println("    • waitForDeployment   - Wait for rollout completion")
	fmt.Println()
	fmt.Println("  🔧 Configuration Management:")
	fmt.Println("    • setDeploymentImage      - Update container images")
	fmt.Println("    • setDeploymentEnv        - Update environment variables")
	fmt.Println("    • setDeploymentResources  - Update resource limits/requests")
	fmt.Println("    • patchDeployment         - Apply JSON/strategic patches")
	fmt.Println()
	fmt.Println("  🔍 Monitoring & Analysis:")
	fmt.Println("    • getDeploymentEvents    - Get deployment events")
	fmt.Println("    • getDeploymentLogs      - Get logs from all pods")
	fmt.Println("    • getDeploymentMetrics   - Get resource metrics")
	fmt.Println("    • getDeploymentYAML      - Export as YAML")
	fmt.Println()
	fmt.Println("  🌐 Batch Operations:")
	fmt.Println("    • listAllDeployments     - List across all namespaces")
	fmt.Println("    • scaleAllDeployments    - Scale all in namespace")
	fmt.Println()

	// Cluster Overview Section
	fmt.Println("🔴 CLUSTER OVERVIEW")
	fmt.Println("  🌍 Global Operations:")
	fmt.Println("    • getClusterOverview     - Cluster-wide resource overview")
	fmt.Println()

	fmt.Println("🔧 ═══════════════════════════════════════════════════════════════")
	fmt.Printf("📊 TOTAL: %d Tools Available\n", getTotalToolCount())
	fmt.Println("💡 Perfect for KubeSphere-like Dashboard Integration!")
	fmt.Println("🔧 ═══════════════════════════════════════════════════════════════")
}

func getTotalToolCount() int {
	return 42 // Update this count as you add more tools
}
