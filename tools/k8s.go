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

// ========== EXTENDED DEPLOYMENT TOOLS ==========

// GetDeploymentEventsTool creates a tool for getting deployment-related events
func GetDeploymentEventsTool() mcp.Tool {
	return mcp.NewTool(
		"getDeploymentEvents",
		mcp.WithDescription("Get events related to a specific deployment for debugging and monitoring"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of events to return (default: 50)")),
	)
}

// GetDeploymentLogsTool creates a tool for getting logs from all pods in a deployment
func GetDeploymentLogsTool() mcp.Tool {
	return mcp.NewTool(
		"getDeploymentLogs",
		mcp.WithDescription("Get logs from all pods in a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithString("container", mcp.Description("Specific container name (optional)")),
		mcp.WithNumber("lines", mcp.Description("Number of lines to retrieve (default: 100)")),
		mcp.WithBoolean("follow", mcp.Description("Follow log output (default: false)")),
	)
}

// RestartDeploymentTool creates a tool for restarting a deployment
func RestartDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restartDeployment",
		mcp.WithDescription("Restart a deployment by triggering a rollout (useful for config reloads)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment to restart")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// WaitForDeploymentTool creates a tool for waiting for deployment to reach desired state
func WaitForDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"waitForDeployment",
		mcp.WithDescription("Wait for a deployment to reach its desired state (ready)"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithNumber("timeout", mcp.Description("Timeout in seconds (default: 300)")),
	)
}

// SetDeploymentImageTool creates a tool for updating container images
func SetDeploymentImageTool() mcp.Tool {
	return mcp.NewTool(
		"setDeploymentImage",
		mcp.WithDescription("Update container image in a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("container", mcp.Required(), mcp.Description("The name of the container to update")),
		mcp.WithString("image", mcp.Required(), mcp.Description("The new container image")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// SetDeploymentEnvTool creates a tool for updating environment variables
func SetDeploymentEnvTool() mcp.Tool {
	return mcp.NewTool(
		"setDeploymentEnv",
		mcp.WithDescription("Update environment variables in a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("container", mcp.Required(), mcp.Description("The name of the container to update")),
		mcp.WithString("env", mcp.Required(), mcp.Description("Environment variables as JSON object (e.g., '{\"KEY1\":\"value1\",\"KEY2\":\"value2\"}')")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// PatchDeploymentTool creates a tool for applying JSON patches
func PatchDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"patchDeployment",
		mcp.WithDescription("Apply a JSON patch to a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("patch", mcp.Required(), mcp.Description("JSON patch to apply")),
		mcp.WithString("patchType", mcp.Description("Type of patch: 'json', 'merge', or 'strategic' (default: 'strategic')")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// GetDeploymentYAMLTool creates a tool for exporting deployment as YAML
func GetDeploymentYAMLTool() mcp.Tool {
	return mcp.NewTool(
		"getDeploymentYAML",
		mcp.WithDescription("Export deployment configuration as YAML"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
		mcp.WithBoolean("export", mcp.Description("Export for backup (removes cluster-specific fields) (default: false)")),
	)
}

// SetDeploymentResourcesTool creates a tool for updating resource requests/limits
func SetDeploymentResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"setDeploymentResources",
		mcp.WithDescription("Update resource requests and limits for a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("container", mcp.Required(), mcp.Description("The name of the container to update")),
		mcp.WithString("resources", mcp.Required(), mcp.Description("Resources as JSON object (e.g., '{\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"},\"limits\":{\"cpu\":\"500m\",\"memory\":\"256Mi\"}}')")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// GetDeploymentMetricsTool creates a tool for getting deployment metrics
func GetDeploymentMetricsTool() mcp.Tool {
	return mcp.NewTool(
		"getDeploymentMetrics",
		mcp.WithDescription("Get CPU and memory metrics for a deployment"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the deployment")),
		mcp.WithString("namespace", mcp.Description("The namespace of the deployment (default: 'default')")),
	)
}

// ListAllDeploymentsTool creates a tool for listing deployments across all namespaces
func ListAllDeploymentsTool() mcp.Tool {
	return mcp.NewTool(
		"listAllDeployments",
		mcp.WithDescription("List deployments across all namespaces with summary information"),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter deployments")),
		mcp.WithBoolean("includeSystem", mcp.Description("Include system namespaces (default: false)")),
	)
}

// ScaleAllDeploymentsTool creates a tool for scaling all deployments in a namespace
func ScaleAllDeploymentsTool() mcp.Tool {
	return mcp.NewTool(
		"scaleAllDeployments",
		mcp.WithDescription("Scale all deployments in a namespace to specified replicas"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to scale deployments in")),
		mcp.WithNumber("replicas", mcp.Required(), mcp.Description("The desired number of replicas for all deployments")),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter which deployments to scale")),
		mcp.WithBoolean("dryRun", mcp.Description("Perform a dry run without making changes (default: false)")),
	)
}

// ========== ADDITIONAL NAMESPACE TOOLS FOR KUBESPHERE-LIKE INTERFACE ==========

// GetNamespaceResourceUsageTool creates a tool for getting resource usage across a namespace
func GetNamespaceResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"getNamespaceResourceUsage",
		mcp.WithDescription("Get resource usage summary for a namespace (pods, deployments, services, etc.)"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace to analyze")),
		mcp.WithBoolean("includeMetrics", mcp.Description("Include CPU/Memory metrics if available (default: false)")),
	)
}

// GetClusterOverviewTool creates a tool for getting cluster-wide overview
func GetClusterOverviewTool() mcp.Tool {
	return mcp.NewTool(
		"getClusterOverview",
		mcp.WithDescription("Get cluster-wide overview including nodes, namespaces, and resource counts"),
		mcp.WithBoolean("includeMetrics", mcp.Description("Include resource metrics if available (default: false)")),
	)
}

// ========== ADDITIONAL POD TOOLS FOR KUBESPHERE-LIKE INTERFACE ==========

// GetPodResourceUsageTool creates a tool for getting pod resource usage
func GetPodResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"getPodResourceUsage",
		mcp.WithDescription("Get resource usage (CPU/Memory) for a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the pod")),
		mcp.WithString("namespace", mcp.Description("The namespace of the pod (default: 'default')")),
	)
}

// GetPodsHealthStatusTool creates a tool for getting health status of pods
func GetPodsHealthStatusTool() mcp.Tool {
	return mcp.NewTool(
		"getPodsHealthStatus",
		mcp.WithDescription("Get health status overview of all pods in a namespace"),
		mcp.WithString("namespace", mcp.Description("The namespace to check (default: 'default')")),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter pods")),
	)
}

// ========== SERVICE TOOLS ==========

// ListServicesTool creates a tool for listing services in a namespace
func ListServicesTool() mcp.Tool {
	return mcp.NewTool(
		"listServices",
		mcp.WithDescription("List all services in a Kubernetes namespace with detailed information"),
		mcp.WithString("namespace", mcp.Description("The namespace to list services from (default: 'default')")),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter services (e.g., 'app=nginx,tier=frontend')")),
	)
}

// GetServiceTool creates a tool for getting detailed information about a specific service
func GetServiceTool() mcp.Tool {
	return mcp.NewTool(
		"getService",
		mcp.WithDescription("Get detailed information about a specific service including endpoints"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// CreateServiceTool creates a tool for creating a new service
func CreateServiceTool() mcp.Tool {
	return mcp.NewTool(
		"createService",
		mcp.WithDescription("Create a new service from a JSON manifest"),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The service manifest in JSON format")),
		mcp.WithString("namespace", mcp.Description("The namespace to create the service in (default: 'default')")),
	)
}

// UpdateServiceTool creates a tool for updating service configurations
func UpdateServiceTool() mcp.Tool {
	return mcp.NewTool(
		"updateService",
		mcp.WithDescription("Update an existing service with new specifications"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service to update")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The updated service manifest in JSON format")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// DeleteServiceTool creates a tool for deleting a service
func DeleteServiceTool() mcp.Tool {
	return mcp.NewTool(
		"deleteService",
		mcp.WithDescription("Delete a service"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service to delete")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// GetServiceEndpointsTool creates a tool for getting service endpoints
func GetServiceEndpointsTool() mcp.Tool {
	return mcp.NewTool(
		"getServiceEndpoints",
		mcp.WithDescription("Get endpoints for a specific service showing backend pods"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// TestServiceConnectivityTool creates a tool for testing service connectivity
func TestServiceConnectivityTool() mcp.Tool {
	return mcp.NewTool(
		"testServiceConnectivity",
		mcp.WithDescription("Test service connectivity and DNS resolution within the cluster"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service to test")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
		mcp.WithNumber("port", mcp.Description("Specific port to test (optional)")),
		mcp.WithString("protocol", mcp.Description("Protocol to test: TCP, UDP (default: TCP)")),
	)
}

// ========== EXTENDED SERVICE TOOLS ==========

// GetServiceEventsTool creates a tool for getting service-related events
func GetServiceEventsTool() mcp.Tool {
	return mcp.NewTool(
		"getServiceEvents",
		mcp.WithDescription("Get events related to a specific service for debugging"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of events to return (default: 50)")),
	)
}

// GetServiceYAMLTool creates a tool for exporting service as YAML
func GetServiceYAMLTool() mcp.Tool {
	return mcp.NewTool(
		"getServiceYAML",
		mcp.WithDescription("Export service configuration as YAML"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
		mcp.WithBoolean("export", mcp.Description("Export for backup (removes cluster-specific fields) (default: false)")),
	)
}

// ExposeDeploymentTool creates a tool for exposing a deployment as a service
func ExposeDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"exposeDeployment",
		mcp.WithDescription("Expose a deployment as a service"),
		mcp.WithString("deployment", mcp.Required(), mcp.Description("The name of the deployment to expose")),
		mcp.WithString("serviceName", mcp.Description("Name for the new service (default: deployment name)")),
		mcp.WithNumber("port", mcp.Required(), mcp.Description("Port for the service")),
		mcp.WithNumber("targetPort", mcp.Description("Target port on the pods (default: same as port)")),
		mcp.WithString("serviceType", mcp.Description("Service type: ClusterIP, NodePort, LoadBalancer (default: ClusterIP)")),
		mcp.WithString("namespace", mcp.Description("The namespace (default: 'default')")),
	)
}

// PatchServiceTool creates a tool for applying patches to services
func PatchServiceTool() mcp.Tool {
	return mcp.NewTool(
		"patchService",
		mcp.WithDescription("Apply a JSON patch to a service"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("patch", mcp.Required(), mcp.Description("JSON patch to apply")),
		mcp.WithString("patchType", mcp.Description("Type of patch: 'json', 'merge', or 'strategic' (default: 'strategic')")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// ListAllServicesTool creates a tool for listing services across all namespaces
func ListAllServicesTool() mcp.Tool {
	return mcp.NewTool(
		"listAllServices",
		mcp.WithDescription("List services across all namespaces with summary information"),
		mcp.WithString("labelSelector", mcp.Description("Optional label selector to filter services")),
		mcp.WithBoolean("includeSystem", mcp.Description("Include system namespaces (default: false)")),
	)
}

// GetServiceMetricsTool creates a tool for getting service metrics
func GetServiceMetricsTool() mcp.Tool {
	return mcp.NewTool(
		"getServiceMetrics",
		mcp.WithDescription("Get service metrics including connection counts and traffic"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// GetServiceTopologyTool creates a tool for getting service topology
func GetServiceTopologyTool() mcp.Tool {
	return mcp.NewTool(
		"getServiceTopology",
		mcp.WithDescription("Get service topology showing relationships with pods and deployments"),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the service")),
		mcp.WithString("namespace", mcp.Description("The namespace of the service (default: 'default')")),
	)
}

// CreateServiceFromPodsTool creates a tool for creating services from pod selectors
func CreateServiceFromPodsTool() mcp.Tool {
	return mcp.NewTool(
		"createServiceFromPods",
		mcp.WithDescription("Create a service that selects specific pods"),
		mcp.WithString("serviceName", mcp.Required(), mcp.Description("Name for the new service")),
		mcp.WithString("labelSelector", mcp.Required(), mcp.Description("Label selector to match pods (e.g., 'app=nginx')")),
		mcp.WithNumber("port", mcp.Required(), mcp.Description("Port for the service")),
		mcp.WithNumber("targetPort", mcp.Description("Target port on the pods (default: same as port)")),
		mcp.WithString("serviceType", mcp.Description("Service type: ClusterIP, NodePort, LoadBalancer (default: ClusterIP)")),
		mcp.WithString("namespace", mcp.Description("The namespace (default: 'default')")),
	)
}
