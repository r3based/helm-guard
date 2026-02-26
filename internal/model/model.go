package model

import "k8s.io/apimachinery/pkg/api/resource"

type Model struct {
	Workloads              []Workload
	Services               []Service
	Ingresses              []Ingress
	ConfigMaps             []ConfigMap
	Secrets                []Secret
	PodDisruptionBudgets   []PodDisruptionBudget
	PersistentVolumeClaims []PersistentVolumeClaim
	HorizontalPodAutoscalers []HorizontalPodAutoscaler
	Roles                    []Role
	ClusterRoles             []ClusterRole
	RoleBindings             []RoleBinding
	ClusterRoleBindings      []ClusterRoleBinding
	CronJobs                 []CronJob
	Jobs                     []Job
	NetworkPolicies          []NetworkPolicy
	Summary                  Summary
	Other                    []ObjectRef // unhandled kinds: kind, apiVersion, name, namespace
}

// ObjectRef is a minimal reference to any Kubernetes object.
type ObjectRef struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
}

// PodSecurityContext holds pod-level security context (spec.template.spec.securityContext).
type PodSecurityContext struct {
	RunAsUser    *int64
	RunAsGroup   *int64
	RunAsNonRoot *bool
	FsGroup      *int64
	SeccompProfile *SeccompProfile
}

type SeccompProfile struct {
	Type             string // RuntimeDefault, Localhost, Unconfined
	LocalhostProfile string
}

type Workload struct {
	Kind               string
	Name               string
	Namespace          string
	Replicas           int64 // 0 for DaemonSet (interpret as N/A)
	PodLabels          map[string]string
	ServiceAccountName string
	HostNetwork        bool
	HostPID            bool
	HostIPC            bool
	PodSecurityContext *PodSecurityContext
	Volumes            []Volume
	InitContainers     []Container
	Containers         []Container
}

// Volume represents a pod-level volume (name + source refs for dependency analysis).
type Volume struct {
	Name          string
	Type          string // configMap, secret, persistentVolumeClaim, emptyDir, hostPath, etc.
	SecretName    string
	ConfigMapName string
	ClaimName     string
}

// VolumeMount is a container volume mount.
type VolumeMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type Ingress struct {
	Name             string
	Namespace        string
	IngressClassName string
	Hosts            []string
	Services         []string   // backend service names
	TLS              []IngressTLS
}

type IngressTLS struct {
	Hosts      []string
	SecretName string
}

type Container struct {
	Name                    string
	Image                   string
	ImagePullPolicy         string // Always, IfNotPresent, "" (default)
	CPURequest              *resource.Quantity
	CPULimit                *resource.Quantity
	MemRequest              *resource.Quantity
	MemLimit                *resource.Quantity
	HasReadiness            bool
	HasLiveness             bool
	HasStartup              bool
	Privileged              bool
	ReadOnlyRootFilesystem  bool
	AllowPrivilegeEscalation *bool // nil = not set (defaults to true when running as root)
	RunAsNonRoot            *bool
	VolumeMounts            []VolumeMount
	// Refs from env/envFrom for dependency analysis (names only, same namespace as pod).
	ConfigMapRefs []string
	SecretRefs    []string
}

type Service struct {
	Name       string
	Namespace  string
	Type       string
	ClusterIP  string // "" or "None" for headless
	Selector   map[string]string
	Ports      []ServicePort
}

type ConfigMap struct {
	Name      string
	Namespace string
	DataKeys  []string
}

type Secret struct {
	Name      string
	Namespace string
	Type      string
}

type PodDisruptionBudget struct {
	Name             string
	Namespace        string
	Selector         map[string]string
	MinAvailable     string // int or percentage string
	MaxUnavailable   string
}

type PersistentVolumeClaim struct {
	Name         string
	Namespace    string
	StorageClass string
	AccessModes  []string
	Capacity     string
}

// HorizontalPodAutoscaler references a scale target (Deployment/StatefulSet etc.) and defines min/max replicas.
type HorizontalPodAutoscaler struct {
	Name       string
	Namespace  string
	TargetKind             string // Deployment, StatefulSet, etc.
	TargetName             string
	MinReplicas            int32
	MaxReplicas            int32
	TargetCPUUtilization   *int32 // percentage, nil if not set
}

// PolicyRule is one rule in Role/ClusterRole (apiGroups, resources, verbs).
type PolicyRule struct {
	APIGroups []string
	Resources []string
	Verbs     []string
}

type Role struct {
	Name      string
	Namespace string
	Rules     []PolicyRule
}

type ClusterRole struct {
	Name  string
	Rules []PolicyRule
}

// Subject is a RoleBinding/ClusterRoleBinding subject (user, group, or service account).
type Subject struct {
	Kind      string
	Name      string
	Namespace string
}

type RoleBinding struct {
	Name      string
	Namespace string
	RoleRef   string // name of Role
	Subjects  []Subject
}

type ClusterRoleBinding struct {
	Name     string
	RoleRef  string // name of ClusterRole
	Subjects []Subject
}

// PodTemplateRef is a minimal pod template (labels + containers) for Job/CronJob.
type PodTemplateRef struct {
	Labels     map[string]string
	Containers []Container
}

type CronJob struct {
	Name        string
	Namespace   string
	Schedule    string
	JobTemplate *PodTemplateRef
}

type Job struct {
	Name       string
	Namespace  string
	PodTemplate *PodTemplateRef
}

// NetworkPolicy peer (podSelector, namespaceSelector, or ipBlock).
type NPPeer struct {
	PodSelector       map[string]string
	NamespaceSelector map[string]string
	IPBlockCIDR       string
}

type NPPort struct {
	Protocol string
	Port     string
}

type NetworkPolicyIngressRule struct {
	Ports []NPPort
	From  []NPPeer
}

type NetworkPolicyEgressRule struct {
	Ports []NPPort
	To    []NPPeer
}

type NetworkPolicy struct {
	Name         string
	Namespace    string
	PodSelector  map[string]string
	PolicyTypes  []string
	IngressRules []NetworkPolicyIngressRule
	EgressRules  []NetworkPolicyEgressRule
}

type ServicePort struct {
	Name       string
	Port       int64
	TargetPort string
	Protocol   string
}

type Summary struct {
	CPURequests resource.Quantity
	CPULimits   resource.Quantity
	MemRequests resource.Quantity
	MemLimits   resource.Quantity
}
