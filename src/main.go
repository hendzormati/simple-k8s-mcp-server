package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
)

var k8sClient *k8s.Client

func main() {
    fmt.Println("Starting simple K8s MCP server...")

    // Initialize Kubernetes client
    var err error
    k8sClient, err = k8s.NewClient()
    if err != nil {
        log.Printf("Warning: Failed to create K8s client: %v", err)
        log.Println("Server will start but K8s features won't work")
    } else {
        // Test connection
        if err := k8sClient.TestConnection(); err != nil {
            log.Printf("Warning: Cannot connect to K8s cluster: %v", err)
        } else {
            fmt.Println("✅ Successfully connected to Kubernetes cluster!")
        }
    }

    // Health check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Server is running!")
    })

    // K8s pods endpoint
    http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
        if k8sClient == nil {
            http.Error(w, "Kubernetes client not available", http.StatusServiceUnavailable)
            return
        }

        pods, err := k8sClient.GetPods()
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get pods: %v", err), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "pods": pods,
            "count": len(pods),
        })
    })

    log.Println("Server starting on port 8080")
    log.Println("Endpoints:")
    log.Println("  - GET /health - Health check")
    log.Println("  - GET /pods   - List pods in default namespace")
    
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}