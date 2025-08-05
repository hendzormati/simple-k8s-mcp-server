package handlers

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/hendzormati/simple-k8s-mcp-server/pkg/k8s"
)

// ListPods returns a handler function for the listPods tool
func ListPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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