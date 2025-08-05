package k8s

import (
    "context"
    "fmt"
    "path/filepath"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
    clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
    // Try to get kubeconfig from home directory
    var kubeconfig string
    if home := homedir.HomeDir(); home != "" {
        kubeconfig = filepath.Join(home, ".kube", "config")
    }

    // Build config from kubeconfig file
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create config: %v", err)
    }

    // Create clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %v", err)
    }

    return &Client{clientset: clientset}, nil
}

// GetPods returns all pods in the default namespace
func (c *Client) GetPods() ([]string, error) {
    pods, err := c.clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get pods: %v", err)
    }

    var podNames []string
    for _, pod := range pods.Items {
        podNames = append(podNames, pod.Name)
    }

    return podNames, nil
}

// TestConnection tests if we can connect to the cluster
func (c *Client) TestConnection() error {
    _, err := c.clientset.Discovery().ServerVersion()
    if err != nil {
        return fmt.Errorf("failed to connect to cluster: %v", err)
    }
    return nil
}