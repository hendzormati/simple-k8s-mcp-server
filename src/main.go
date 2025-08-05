package main

import (
    "fmt"
    "log"

    "github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
    "github.com/hendzormati/simple-k8s-mcp-server/tools"
    "github.com/hendzormati/simple-k8s-mcp-server/handlers"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    fmt.Println("Starting simple K8s MCP server...")

    // Initialize Kubernetes client
    k8sClient, err := k8s.NewClient()
    if err != nil {
        log.Fatalf("Failed to create K8s client: %v", err)
    }

    // Test connection
    if err := k8sClient.TestConnection(); err != nil {
        log.Printf("Warning: Cannot connect to K8s cluster: %v", err)
        log.Println("Server will start but K8s features won't work")
    } else {
        fmt.Println("✅ Successfully connected to Kubernetes cluster!")
    }

    // Create MCP server
    mcpServer := server.NewMCPServer(
        "Simple K8s MCP Server",
        "1.0.0",
    )

    // Register our first tool
    mcpServer.AddTool(tools.ListPodsTool(), handlers.ListPods(k8sClient))

    fmt.Println("MCP Server initialized with tools:")
    fmt.Println("  - listPods: List pods in a namespace")
    
    // Start server in stdio mode (for MCP protocol)
    fmt.Println("Starting MCP server in stdio mode...")
    if err := server.ServeStdio(mcpServer); err != nil {
        log.Fatalf("Failed to start MCP server: %v", err)
    }
}