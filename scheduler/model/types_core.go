package model

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type ObjectMeta struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Labels            map[string]string `json:"labels"`
	UID               string            `json:"uid"`
	Kind              string            `json:"kind"`
	ResourceVersion   string            `json:"resource_version"`
	Generation        int64             `json:"generation"`
	CreationTimestamp time.Time         `json:"creation_timestamp"`
	DeletionTimestamp *time.Time        `json:"deletion_timestamp"`
}

type JvirtMeta struct {
	ResourceId string `json:"resource_id"`
	InstanceId string `json:"instance_id"`
	AllocId    string `json:"alloc_id"`
	HostId     string `json:"host_id"`
	TaskId     string `json:"task_id"`
}

type ObjectReference struct {
	Kind            string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	Namespace       string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	Name            string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
	UID             string `json:"uid,omitempty" protobuf:"bytes,4,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`
}

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
	PodUnknown   PodPhase = "Unknown"
)

type PodConditionType string

const (
	ContainersReady PodConditionType = "ContainersReady"
	PodInitialized  PodConditionType = "Initialized"
	PodReady        PodConditionType = "Ready"
	PodScheduled    PodConditionType = "PodScheduled"
)

const (
	PodReasonUnschedulable = "Unschedulable"
)

type Pod struct {
	metav1.TypeMeta `json:"meta"`
	ObjectMeta      `json:"metadata"`
	JvirtMeta       `json:"jvirt_meta"`
	Spec            PodSpec   `json:"spec"`
	Status          PodStatus `json:"status"`
}

type PodStatus struct {
	StartTime         *time.Time `json:"start_time"`
	Phase             PodPhase   `json:"phase"`
	NominatedNodeName string     `json:"dominated_node_name"`
}

type PodSpec struct {
	Containers   []Container       `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	NodeName                  string                     `json:"nodeName,omitempty" protobuf:"bytes,10,opt,name=nodeName"`
	HostPID                   bool                       `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`
	Hostname                  string                     `json:"hostname,omitempty" protobuf:"bytes,16,opt,name=hostname"`
	Subdomain                 string                     `json:"subdomain,omitempty" protobuf:"bytes,17,opt,name=subdomain"`
	Affinity                  *Affinity                  `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	SchedulerName             string                     `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`
	Tolerations               []Toleration               `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	HostAliases               []HostAlias                `json:"hostAliases,omitempty" patchStrategy:"merge" patchMergeKey:"ip" protobuf:"bytes,23,rep,name=hostAliases"`
	PriorityClassName         string                     `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
	Priority                  *int32                     `json:"priority,omitempty" protobuf:"bytes,25,opt,name=priority"`
	PreemptionPolicy          *PreemptionPolicy          `json:"preemptionPolicy,omitempty" protobuf:"bytes,31,opt,name=preemptionPolicy"`
	Overhead                  ResourceList               `json:"overhead,omitempty" protobuf:"bytes,32,opt,name=overhead"`
	TopologySpreadConstraints []TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty" patchStrategy:"merge" patchMergeKey:"topologyKey" protobuf:"bytes,33,opt,name=topologySpreadConstraints"`
	SetHostnameAsFQDN         *bool                      `json:"setHostnameAsFQDN,omitempty" protobuf:"varint,35,opt,name=setHostnameAsFQDN"`
}

type Container struct {
	Name      string               `json:"name" protobuf:"bytes,1,opt,name=name"`
	Image     string               `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`
	Resources ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList,castkey=ResourceName"`
	Requests ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList,castkey=ResourceName"`
}

type Taint struct {
	Key       string      `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value     string      `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	Effect    TaintEffect `json:"effect" protobuf:"bytes,3,opt,name=effect,casttype=TaintEffect"`
	TimeAdded *time.Time  `json:"timeAdded,omitempty" protobuf:"bytes,4,opt,name=timeAdded"`
}

type TaintEffect string

const (
	TaintEffectNoSchedule       TaintEffect = "NoSchedule"
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"

	TaintEffectNoExecute TaintEffect = "NoExecute"
)

type Toleration struct {
	Key               string             `json:"key,omitempty" protobuf:"bytes,1,opt,name=key"`
	Operator          TolerationOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`
	Value             string             `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`
	Effect            TaintEffect        `json:"effect,omitempty" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
	TolerationSeconds *int64             `json:"tolerationSeconds,omitempty" protobuf:"varint,5,opt,name=tolerationSeconds"`
}

type TolerationOperator string

const (
	TolerationOpExists TolerationOperator = "Exists"
	TolerationOpEqual  TolerationOperator = "Equal"
)

type HostAlias struct {
	IP        string   `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`
	Hostnames []string `json:"hostnames,omitempty" protobuf:"bytes,2,rep,name=hostnames"`
}

type PreemptionPolicy string

const (
	PreemptLowerPriority PreemptionPolicy = "PreemptLowerPriority"
	PreemptNever         PreemptionPolicy = "Never"
)

type Node struct {
	metav1.TypeMeta
	ObjectMeta
	Spec   NodeSpec   `json:"spec"`
	Status NodeStatus `json:"status"`
}
type NodeSpec struct {
	Taints        []Taint `json:"taints"`
	Unschedulable bool    `json:"unschedulable"`
}

type NodeStatus struct {
	Capacity    ResourceList    `json:"capacity,omitempty" protobuf:"bytes,1,rep,name=capacity,casttype=ResourceList,castkey=ResourceName"`
	Allocatable ResourceList    `json:"allocatable,omitempty" protobuf:"bytes,2,rep,name=allocatable,casttype=ResourceList,castkey=ResourceName"`
	Phase       NodePhase       `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=NodePhase"`
	Conditions  []NodeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,4,rep,name=conditions"`
}

type Binding struct {
	ObjectMeta `json:"metadata"`
	Target     ObjectReference `json:"target"`
}

type NodePhase string

const (
	NodePending    NodePhase = "Pending"
	NodeRunning    NodePhase = "Running"
	NodeTerminated NodePhase = "Terminated"
)

type NodeConditionType string

const (
	NodeReady              NodeConditionType = "Ready"
	NodeMemoryPressure     NodeConditionType = "MemoryPressure"
	NodeDiskPressure       NodeConditionType = "DiskPressure"
	NodePIDPressure        NodeConditionType = "PIDPressure"
	NodeNetworkUnavailable NodeConditionType = "NetworkUnavailable"
)

type NodeCondition struct {
	Type               NodeConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=NodeConditionType"`
	Status             ConditionStatus   `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	LastHeartbeatTime  time.Time         `json:"lastHeartbeatTime,omitempty" protobuf:"bytes,3,opt,name=lastHeartbeatTime"`
	LastTransitionTime time.Time         `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason             string            `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	Message            string            `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type NodeList struct {
	Items []Node `json:"items"`
}

type ResourceList map[ResourceName]resource.Quantity

type ResourceName string

const (
	ResourceCPU              ResourceName = "cpu"
	ResourceMemory           ResourceName = "memory"
	ResourceStorage          ResourceName = "storage"
	ResourceEphemeralStorage ResourceName = "ephemeral-storage"
)

const (
	ResourcePods                     ResourceName = "pods"
	ResourceQuotas                   ResourceName = "resourcequotas"
	ResourceRequestsCPU              ResourceName = "requests.cpu"
	ResourceRequestsMemory           ResourceName = "requests.memory"
	ResourceRequestsStorage          ResourceName = "requests.storage"
	ResourceRequestsEphemeralStorage ResourceName = "requests.ephemeral-storage"
	ResourceLimitsCPU                ResourceName = "limits.cpu"
	ResourceLimitsMemory             ResourceName = "limits.memory"
	ResourceLimitsEphemeralStorage   ResourceName = "limits.ephemeral-storage"
)

type UnsatisfiableConstraintAction string

const (
	DoNotSchedule  UnsatisfiableConstraintAction = "DoNotSchedule"
	ScheduleAnyway UnsatisfiableConstraintAction = "ScheduleAnyway"
)

type TopologySpreadConstraint struct {
	MaxSkew           int32                         `json:"max_skew"`
	TopologyKey       string                        `json:"topology_key"`
	WhenUnsatisfiable UnsatisfiableConstraintAction `json:"when_unsatisfiable"`
	LabelSelector     *LabelSelector                `json:"label_selector"`
}

const (
	DefaultSchedulerName = "default-scheduler"
)

type PodCondition struct {
	Type               PodConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=PodConditionType"`
	Status             ConditionStatus  `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	LastProbeTime      time.Time        `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	LastTransitionTime time.Time        `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason             string           `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	Message            string           `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type Capability string

type Capabilities struct {
	Add  []Capability `json:"add,omitempty" protobuf:"bytes,1,rep,name=add,casttype=Capability"`
	Drop []Capability `json:"drop,omitempty" protobuf:"bytes,2,rep,name=drop,casttype=Capability"`
}
