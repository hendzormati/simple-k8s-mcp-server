package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

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

// GetNamespaceYAMLTool creates a tool for getting namespace YAML definition
func GetNamespaceYAMLTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceYAML",
		mcp.WithDescription("Get the YAML definition of a namespace"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to get YAML for")),
	)
}

// SetNamespaceResourceQuotaTool creates a tool for setting resource quota
func SetNamespaceResourceQuotaTool() mcp.Tool {
	return mcp.NewTool(
		"setNamespaceResourceQuota",
		mcp.WithDescription("Create or update a resource quota in a namespace"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to set the resource quota in")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The resource quota manifest in JSON format (e.g., '{\"apiVersion\":\"v1\",\"kind\":\"ResourceQuota\",\"metadata\":{\"name\":\"my-quota\"},\"spec\":{\"hard\":{\"requests.cpu\":\"1\",\"requests.memory\":\"1Gi\"}}}')")),
	)
}

// GetNamespaceLimitRangesTool creates a tool for getting limit ranges
func GetNamespaceLimitRangesTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceLimitRanges",
		mcp.WithDescription("Get limit ranges for a specific namespace"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to get limit ranges from")),
	)
}

// SetNamespaceLimitRangeTool creates a tool for setting limit range
func SetNamespaceLimitRangeTool() mcp.Tool {
	return mcp.NewTool(
		"setNamespaceLimitRange",
		mcp.WithDescription("Create or update a limit range in a namespace"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to set the limit range in")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The limit range manifest in JSON format (e.g., '{\"apiVersion\":\"v1\",\"kind\":\"LimitRange\",\"metadata\":{\"name\":\"my-limit-range\"},\"spec\":{\"limits\":[{\"type\":\"Container\",\"default\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}]}}')")),
	)
}

// SmartDeleteNamespaceTool creates a tool for intelligent namespace deletion
func SmartDeleteNamespaceTool() mcp.Tool {
	return mcp.NewTool(
		"smartDeleteNamespace",
		mcp.WithDescription("Intelligently delete a namespace using the best strategy for the cluster type"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the namespace to delete")),
		mcp.WithBoolean("force", mcp.Description("Force delete if regular deletion fails (default: true)")),
	)
}

// ========== POD TOOLS ==========

// ListPodsTool creates a tool for listing pods in a namespace
func ListPodsTool() mcp.Tool {
	return mcp.NewTool(
		"listPods",
		mcp.WithDescription("List all pods in a Kubernetes namespace with detailed information"),
		mcp.WithString("namespace", mcp.Description("The namespace to list pods from (default: 'default')")),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter pods (e.g., 'app=nginx,version=v1')")),
	)
}

// GetPodTool creates a tool for getting detailed information about a specific pod
func GetPodTool() mcp.Tool {
	return mcp.NewTool(
		"getPod",
		mcp.WithDescription("Get detailed information about a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
	)
}

// GetPodLogsTool creates a tool for getting pod logs
func GetPodLogsTool() mcp.Tool {
	return mcp.NewTool(
		"getPodLogs",
		mcp.WithDescription("Get logs from a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithString("containerName", mcp.Description("Optional container name (if pod has multiple containers)")),
		mcp.WithNumber("tailLines", mcp.Description("Number of lines to tail from the end of logs (default: 100)")),
		mcp.WithBoolean("follow", mcp.Description("Follow log output (stream logs)")),
		mcp.WithBoolean("previous", mcp.Description("Get logs from previous container instance")),
	)
}

// GetPodMetricsTool creates a tool for getting pod resource metrics
func GetPodMetricsTool() mcp.Tool {
	return mcp.NewTool(
		"getPodMetrics",
		mcp.WithDescription("Get CPU and memory metrics for a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
	)
}

// DescribePodTool creates a tool for describing a pod (like kubectl describe)
func DescribePodTool() mcp.Tool {
	return mcp.NewTool(
		"describePod",
		mcp.WithDescription("Get comprehensive description of a pod including events and status"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
	)
}

// DeletePodTool creates a tool for deleting a pod
func DeletePodTool() mcp.Tool {
	return mcp.NewTool(
		"deletePod",
		mcp.WithDescription("Delete a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod to delete")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithNumber("gracePeriodSeconds", mcp.Description("Grace period for pod termination (default: 30)")),
	)
}

// GetPodEventsTool creates a tool for getting events related to a pod
func GetPodEventsTool() mcp.Tool {
	return mcp.NewTool(
		"getPodEvents",
		mcp.WithDescription("Get events related to a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
	)
}

// RestartPodTool creates a tool for restarting a pod (by deleting it)
func RestartPodTool() mcp.Tool {
	return mcp.NewTool(
		"restartPod",
		mcp.WithDescription("Restart a pod by deleting it (useful for pods managed by deployments)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod to restart")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
	)
}

// CreatePodTool creates a tool for creating a new pod
func CreatePodTool() mcp.Tool {
	return mcp.NewTool(
		"createPod",
		mcp.WithDescription("Create a new pod from a JSON manifest"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace where the pod will be created")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The pod manifest in JSON format (e.g., '{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"my-pod\"},\"spec\":{\"containers\":[{\"name\":\"nginx\",\"image\":\"nginx:latest\"}]}}')")),
	)
}

// UpdatePodTool creates a tool for updating pod metadata
func UpdatePodTool() mcp.Tool {
	return mcp.NewTool(
		"updatePod",
		mcp.WithDescription("Update pod labels and annotations (Note: Pod specs are generally immutable after creation)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod to update")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithString("labels", mcp.Description("Optional labels to add/update in JSON format (e.g., '{\"env\":\"prod\",\"version\":\"v2\"}')")),
		mcp.WithString("annotations", mcp.Description("Optional annotations to add/update in JSON format (e.g., '{\"description\":\"Updated pod\",\"owner\":\"team-a\"}')")),
	)
}

// ========== DEPLOYMENT TOOLS ==========

// ListDeploymentsTool creates a tool for listing deployments in a namespace
func ListDeploymentsTool() mcp.Tool {
	return mcp.NewTool(
		"listDeployments",
		mcp.WithDescription("List all deployments in a Kubernetes namespace with detailed information"),
		mcp.WithString("namespace", mcp.Description("The namespace to list deployments from (default: 'default')")),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter deployments (e.g., 'app=nginx,version=v1')")),
	)
}

// GetDeploymentTool creates a tool for getting detailed information about a specific deployment
func GetDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"getDeployment",
		mcp.WithDescription("Get detailed information about a specific deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// CreateDeploymentTool creates a tool for creating a new deployment
func CreateDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"createDeployment",
		mcp.WithDescription("Create a new deployment from a JSON manifest"),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The deployment manifest in JSON format")),
		mcp.WithString("namespace", mcp.Description("The namespace to create the deployment in (default: 'default')")),
	)
}

// UpdateDeploymentTool creates a tool for updating deployment specifications
func UpdateDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"updateDeployment",
		mcp.WithDescription("Update an existing deployment with new specifications"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to update")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The updated deployment manifest in JSON format")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// DeleteDeploymentTool creates a tool for deleting a deployment
func DeleteDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"deleteDeployment",
		mcp.WithDescription("Delete a deployment and optionally its replica sets and pods"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to delete")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithBoolean("cascade", mcp.Description("Whether to delete associated replica sets and pods (default: true)")),
	)
}

// ScaleDeploymentTool creates a tool for scaling deployment replicas
func ScaleDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"scaleDeployment",
		mcp.WithDescription("Scale a deployment to the specified number of replicas"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to scale")),
		mcp.WithNumber("replicas", mcp.Required(), mcp.Description("The desired number of replicas")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// RolloutStatusTool creates a tool for checking rollout status
func RolloutStatusTool() mcp.Tool {
	return mcp.NewTool(
		"rolloutStatus",
		mcp.WithDescription("Check the rollout status of a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithBoolean("watch", mcp.Description("Whether to watch for status changes (default: false)")),
	)
}

// RolloutHistoryTool creates a tool for getting rollout history
func RolloutHistoryTool() mcp.Tool {
	return mcp.NewTool(
		"rolloutHistory",
		mcp.WithDescription("Get the rollout history of a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithNumber("revision", mcp.Description("Optional specific revision to get details for")),
	)
}

// RolloutUndoTool creates a tool for rolling back deployments
func RolloutUndoTool() mcp.Tool {
	return mcp.NewTool(
		"rolloutUndo",
		mcp.WithDescription("Rollback a deployment to a previous revision"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to rollback")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithNumber("toRevision", mcp.Description("Specific revision to rollback to (default: previous revision)")),
	)
}

// PauseDeploymentTool creates a tool for pausing deployments
func PauseDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"pauseDeployment",
		mcp.WithDescription("Pause a deployment to prevent further rollouts"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to pause")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// ResumeDeploymentTool creates a tool for resuming deployments
func ResumeDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"resumeDeployment",
		mcp.WithDescription("Resume a paused deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to resume")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}
