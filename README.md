# Simple K8s MCP Server

A simplified Kubernetes MCP (Model Context Protocol) server for learning Go and K8s integration. This project provides an easy-to-use interface for managing Kubernetes namespaces and pods through various modes including stdio and SSE.

## Features

### Namespace Management
- **List Namespaces**: Get all namespaces in your cluster with detailed information
- **Get Namespace**: Retrieve comprehensive details about a specific namespace
- **Create Namespace**: Create new namespaces with optional labels and annotations
- **Update Namespace**: Modify labels and annotations of existing namespaces
- **Delete Namespace**: Standard namespace deletion with status monitoring
- **Smart Delete Namespace**: Intelligent deletion that automatically chooses the best strategy
- **Force Delete Namespace**: Enhanced deletion for stuck namespaces using multiple strategies
- **Get Namespace YAML**: Export namespace configuration in YAML format
- **Resource Quota Management**: Set and get resource quotas for namespaces
- **Limit Range Management**: Configure and retrieve limit ranges
- **Namespace Events**: Get events related to specific namespaces

### Pod Management
- **List Pods**: Get all pods in a namespace with filtering options
- **Get Pod**: Retrieve detailed information about specific pods
- **Get Pod Logs**: Stream or retrieve logs from pod containers
- **Get Pod Metrics**: Monitor CPU and memory usage
- **Describe Pod**: Get comprehensive pod descriptions (kubectl describe equivalent)
- **Delete Pod**: Remove specific pods with configurable grace periods
- **Restart Pod**: Restart pods by deletion and recreation
- **Create Pod**: Deploy new pods from JSON manifests
- **Update Pod**: Modify pod labels and annotations
- **Get Pod Events**: Retrieve events related to specific pods

### Multi-Environment Support
- **Auto-Detection**: Automatically detects and configures for different cluster types
- **K3s Integration**: Native support for K3s clusters with enhanced deletion strategies
- **In-Cluster Deployment**: Works seamlessly when deployed as a pod in Kubernetes
- **Standard Kubernetes**: Full compatibility with managed and self-hosted clusters

## Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster (K8s or K3s)
- `kubectl` configured with appropriate cluster access

## Installation & Setup

### 1. Install Go (if not already installed)

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Clone and Build

```bash
git clone https://github.com/hendzormati/simple-k8s-mcp-server.git
cd simple-k8s-mcp-server
go mod tidy
go build -o main src/main.go
```

### 3. Set Up Test Environment (Optional)

```bash
kubectl run test-pod --image=nginx --restart=Never
```

## Cluster Setup

### For K3s Cluster

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.32.6+k3s1" \
  sh -s - \
  --write-kubeconfig-mode 644
```

The server automatically detects K3s installations and applies appropriate optimizations.

### For Standard Kubernetes

Ensure your `~/.kube/config` is properly configured or set the `KUBECONFIG` environment variable.

## Usage

### Stdio Mode

Start the server in stdio mode for direct CLI interaction:

```bash
./main --mode stdio
```

#### Example Commands:

**Create a namespace:**
```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"createNamespace","arguments":{"name":"test-interactive","labels":"{}","annotations":"{}"}}}
```

**Get namespace details:**
```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"getNamespace","arguments":{"name":"test-interactive"}}}
```

**Delete namespace:**
```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"deleteNamespace","arguments":{"name":"test-interactive"}}}
```

**Smart delete (recommended for K3s):**
```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"smartDeleteNamespace","arguments":{"name":"stuck-namespace","force":true}}}
```

### SSE Mode

Start the server in SSE mode for web integration:

```bash
./main --mode sse --host 192.168.1.14 --port 8080
```

The SSE mode works perfectly with automation platforms like n8n workflows, providing a conversational interface for Kubernetes management.

#### Example Conversation Flow:

**User:** "Are you connected to the cluster?"

**Server:** "Yes, I am connected to the Kubernetes cluster and ready to help you manage your resources."

**User:** "Create a namespace named ns-delete"

**Server:** "Namespace 'ns-delete' created successfully."

**User:** "Now describe it for me"

**Server:** "The namespace 'ns-delete' was created on 2025-08-15T18:29:14+01:00, has the label kubernetes.io/metadata.name: ns-delete, and its current status is Active."

**User:** "I want to update it with a label, give me an example"

**Server:** "The namespace 'ns-delete' has been updated with the label 'environment: test'."

**User:** "Now delete it"

**Server:** "WARNING: Deleting the namespace 'ns-delete' will delete all resources within it. Are you sure you want to proceed?"

**User:** "Yes"

**Server:** "Namespace 'ns-delete' is terminating. If it gets stuck, you can use forceDeleteNamespace to remove it."

**User:** "Now list me the namespaces"

**Server:** "The current namespaces are: default, kube-node-lease, kube-public, and kube-system. The ns-delete namespace is no longer listed, indicating it has been successfully deleted."

## Enhanced Deletion Features

This server includes advanced deletion capabilities specifically designed for environments where namespaces can get stuck in "Terminating" state:

- **Smart Delete**: Automatically chooses between regular and force deletion
- **Multiple Strategies**: Uses 4 different approaches for stuck namespaces
- **K3s Optimized**: Special handling for K3s API discovery issues
- **Finalizer Management**: Automatic removal of blocking finalizers

## Testing Environments

This project has been thoroughly tested in:

- **Killerkoda**: Online Kubernetes playground environment
- **K3s Cluster**: Lightweight Kubernetes distribution
- **Standard Kubernetes**: Various managed and self-hosted clusters

## Available Server Modes

- **stdio**: Direct command-line interface for scripting and CLI tools
- **sse**: Server-Sent Events mode for web applications and automation platforms
- **http**: Standard HTTP API mode

## Configuration Options

The server automatically detects your cluster configuration from:
1. In-cluster service account (when running as a pod)
2. `KUBECONFIG` environment variable
3. K3s default locations (`/etc/rancher/k3s/k3s.yaml`)
4. Standard kubeconfig locations (`~/.kube/config`)

## Acknowledgments

This project is inspired by the [k8s-mcp-server](https://github.com/reza-gholizade/k8s-mcp-server) project. While maintaining the core MCP protocol compatibility, this simplified version focuses on learning Go and Kubernetes integration with enhanced namespace and pod management capabilities.

## License

This project is open source and available under the MIT License. See the [LICENSE](LICENSE) file for more information.