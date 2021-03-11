package model

import "time"

type PodDisruptionBudgetSpec struct {
	MinAvailable *string `json:"minAvailable,omitempty" protobuf:"bytes,1,opt,name=minAvailable"`

	Selector *LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	MaxUnavailable *string `json:"maxUnavailable,omitempty" protobuf:"bytes,3,opt,name=maxUnavailable"`
}

type PodDisruptionBudgetStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	DisruptedPods map[string]time.Time `json:"disruptedPods,omitempty" protobuf:"bytes,2,rep,name=disruptedPods"`

	DisruptionsAllowed int32 `json:"disruptionsAllowed" protobuf:"varint,3,opt,name=disruptionsAllowed"`

	CurrentHealthy int32 `json:"currentHealthy" protobuf:"varint,4,opt,name=currentHealthy"`

	DesiredHealthy int32 `json:"desiredHealthy" protobuf:"varint,5,opt,name=desiredHealthy"`

	ExpectedPods int32 `json:"expectedPods" protobuf:"varint,6,opt,name=expectedPods"`
}

type PodDisruptionBudget struct {
	ObjectMeta
	JvirtMeta
	Spec   PodDisruptionBudgetSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status PodDisruptionBudgetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type PodDisruptionBudgetList struct {
	Items []PodDisruptionBudget `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type Eviction struct {
	ObjectMeta

	DeleteOptions *string `json:"deleteOptions,omitempty" protobuf:"bytes,2,opt,name=deleteOptions"`
}

type PodSecurityPolicy struct {
	ObjectMeta

	Spec PodSecurityPolicySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type PodSecurityPolicySpec struct {
	Privileged                      bool                              `json:"privileged,omitempty"`
	DefaultAddCapabilities          []Capability                      `json:"defaultAddCapabilities,omitempty"`
	RequiredDropCapabilities        []Capability                      `json:"requiredDropCapabilities,omitempty"`
	AllowedCapabilities             []Capability                      `json:"allowedCapabilities,omitempty" protobuf:"bytes,4,rep,name=allowedCapabilities,casttype=k8s.io/api/core/v1.Capability"`
	Volumes                         []FSType                          `json:"volumes,omitempty" protobuf:"bytes,5,rep,name=volumes,casttype=FSType"`
	HostNetwork                     bool                              `json:"hostNetwork,omitempty" protobuf:"varint,6,opt,name=hostNetwork"`
	HostPorts                       []HostPortRange                   `json:"hostPorts,omitempty" protobuf:"bytes,7,rep,name=hostPorts"`
	HostPID                         bool                              `json:"hostPID,omitempty" protobuf:"varint,8,opt,name=hostPID"`
	HostIPC                         bool                              `json:"hostIPC,omitempty" protobuf:"varint,9,opt,name=hostIPC"`
	SELinux                         SELinuxStrategyOptions            `json:"seLinux" protobuf:"bytes,10,opt,name=seLinux"`
	RunAsUser                       RunAsUserStrategyOptions          `json:"runAsUser" protobuf:"bytes,11,opt,name=runAsUser"`
	RunAsGroup                      *RunAsGroupStrategyOptions        `json:"runAsGroup,omitempty" protobuf:"bytes,22,opt,name=runAsGroup"`
	SupplementalGroups              SupplementalGroupsStrategyOptions `json:"supplementalGroups" protobuf:"bytes,12,opt,name=supplementalGroups"`
	FSGroup                         FSGroupStrategyOptions            `json:"fsGroup" protobuf:"bytes,13,opt,name=fsGroup"`
	ReadOnlyRootFilesystem          bool                              `json:"readOnlyRootFilesystem,omitempty" protobuf:"varint,14,opt,name=readOnlyRootFilesystem"`
	DefaultAllowPrivilegeEscalation *bool                             `json:"defaultAllowPrivilegeEscalation,omitempty" protobuf:"varint,15,opt,name=defaultAllowPrivilegeEscalation"`
	AllowPrivilegeEscalation        *bool                             `json:"allowPrivilegeEscalation,omitempty" protobuf:"varint,16,opt,name=allowPrivilegeEscalation"`
	AllowedHostPaths                []AllowedHostPath                 `json:"allowedHostPaths,omitempty" protobuf:"bytes,17,rep,name=allowedHostPaths"`
	AllowedFlexVolumes              []AllowedFlexVolume               `json:"allowedFlexVolumes,omitempty" protobuf:"bytes,18,rep,name=allowedFlexVolumes"`
	AllowedCSIDrivers               []AllowedCSIDriver                `json:"allowedCSIDrivers,omitempty" protobuf:"bytes,23,rep,name=allowedCSIDrivers"`
	AllowedUnsafeSysctls            []string                          `json:"allowedUnsafeSysctls,omitempty" protobuf:"bytes,19,rep,name=allowedUnsafeSysctls"`
	ForbiddenSysctls                []string                          `json:"forbiddenSysctls,omitempty" protobuf:"bytes,20,rep,name=forbiddenSysctls"`
	AllowedProcMountTypes           []string                          `json:"allowedProcMountTypes,omitempty" protobuf:"bytes,21,opt,name=allowedProcMountTypes"`
	RuntimeClass                    *RuntimeClassStrategyOptions      `json:"runtimeClass,omitempty" protobuf:"bytes,24,opt,name=runtimeClass"`
}

type AllowedHostPath struct {
	PathPrefix string `json:"pathPrefix,omitempty" protobuf:"bytes,1,rep,name=pathPrefix"`

	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`
}

var AllowAllCapabilities Capability = "*"

type FSType string

const (
	AzureFile             FSType = "azureFile"
	Flocker               FSType = "flocker"
	FlexVolume            FSType = "flexVolume"
	HostPath              FSType = "hostPath"
	EmptyDir              FSType = "emptyDir"
	GCEPersistentDisk     FSType = "gcePersistentDisk"
	AWSElasticBlockStore  FSType = "awsElasticBlockStore"
	GitRepo               FSType = "gitRepo"
	Secret                FSType = "secret"
	NFS                   FSType = "nfs"
	ISCSI                 FSType = "iscsi"
	Glusterfs             FSType = "glusterfs"
	PersistentVolumeClaim FSType = "persistentVolumeClaim"
	RBD                   FSType = "rbd"
	Cinder                FSType = "cinder"
	CephFS                FSType = "cephFS"
	DownwardAPI           FSType = "downwardAPI"
	FC                    FSType = "fc"
	ConfigMap             FSType = "configMap"
	VsphereVolume         FSType = "vsphereVolume"
	Quobyte               FSType = "quobyte"
	AzureDisk             FSType = "azureDisk"
	PhotonPersistentDisk  FSType = "photonPersistentDisk"
	StorageOS             FSType = "storageos"
	Projected             FSType = "projected"
	PortworxVolume        FSType = "portworxVolume"
	ScaleIO               FSType = "scaleIO"
	CSI                   FSType = "csi"
	Ephemeral             FSType = "ephemeral"
	All                   FSType = "*"
)

type AllowedFlexVolume struct {
	Driver string `json:"driver" protobuf:"bytes,1,opt,name=driver"`
}

type AllowedCSIDriver struct {
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
}

type HostPortRange struct {
	Min int32 `json:"min" protobuf:"varint,1,opt,name=min"`
	Max int32 `json:"max" protobuf:"varint,2,opt,name=max"`
}

type SELinuxStrategyOptions struct {
	Rule           SELinuxStrategy `json:"rule" protobuf:"bytes,1,opt,name=rule,casttype=SELinuxStrategy"`
	SELinuxOptions *string         `json:"seLinuxOptions,omitempty" protobuf:"bytes,2,opt,name=seLinuxOptions"`
}

type SELinuxStrategy string

const (
	SELinuxStrategyMustRunAs SELinuxStrategy = "MustRunAs"
	SELinuxStrategyRunAsAny  SELinuxStrategy = "RunAsAny"
)

type RunAsUserStrategyOptions struct {
	Rule   RunAsUserStrategy `json:"rule" protobuf:"bytes,1,opt,name=rule,casttype=RunAsUserStrategy"`
	Ranges []IDRange         `json:"ranges,omitempty" protobuf:"bytes,2,rep,name=ranges"`
}

type RunAsGroupStrategyOptions struct {
	Rule   RunAsGroupStrategy `json:"rule" protobuf:"bytes,1,opt,name=rule,casttype=RunAsGroupStrategy"`
	Ranges []IDRange          `json:"ranges,omitempty" protobuf:"bytes,2,rep,name=ranges"`
}

type IDRange struct {
	Min int64 `json:"min" protobuf:"varint,1,opt,name=min"`
	Max int64 `json:"max" protobuf:"varint,2,opt,name=max"`
}

type RunAsUserStrategy string

const (
	RunAsUserStrategyMustRunAs        RunAsUserStrategy = "MustRunAs"
	RunAsUserStrategyMustRunAsNonRoot RunAsUserStrategy = "MustRunAsNonRoot"
	RunAsUserStrategyRunAsAny         RunAsUserStrategy = "RunAsAny"
)

type RunAsGroupStrategy string

const (
	RunAsGroupStrategyMayRunAs  RunAsGroupStrategy = "MayRunAs"
	RunAsGroupStrategyMustRunAs RunAsGroupStrategy = "MustRunAs"
	RunAsGroupStrategyRunAsAny  RunAsGroupStrategy = "RunAsAny"
)

type FSGroupStrategyOptions struct {
	Rule   FSGroupStrategyType `json:"rule,omitempty" protobuf:"bytes,1,opt,name=rule,casttype=FSGroupStrategyType"`
	Ranges []IDRange           `json:"ranges,omitempty" protobuf:"bytes,2,rep,name=ranges"`
}

type FSGroupStrategyType string

const (
	FSGroupStrategyMayRunAs  FSGroupStrategyType = "MayRunAs"
	FSGroupStrategyMustRunAs FSGroupStrategyType = "MustRunAs"
	FSGroupStrategyRunAsAny  FSGroupStrategyType = "RunAsAny"
)

type SupplementalGroupsStrategyOptions struct {
	Rule   SupplementalGroupsStrategyType `json:"rule,omitempty" protobuf:"bytes,1,opt,name=rule,casttype=SupplementalGroupsStrategyType"`
	Ranges []IDRange                      `json:"ranges,omitempty" protobuf:"bytes,2,rep,name=ranges"`
}

type SupplementalGroupsStrategyType string

const (
	SupplementalGroupsStrategyMayRunAs  SupplementalGroupsStrategyType = "MayRunAs"
	SupplementalGroupsStrategyMustRunAs SupplementalGroupsStrategyType = "MustRunAs"
	SupplementalGroupsStrategyRunAsAny  SupplementalGroupsStrategyType = "RunAsAny"
)

type RuntimeClassStrategyOptions struct {
	AllowedRuntimeClassNames []string `json:"allowedRuntimeClassNames" protobuf:"bytes,1,rep,name=allowedRuntimeClassNames"`
	DefaultRuntimeClassName  *string  `json:"defaultRuntimeClassName,omitempty" protobuf:"bytes,2,opt,name=defaultRuntimeClassName"`
}

const AllowAllRuntimeClassNames = "*"

type PodSecurityPolicyList struct {
	Items []PodSecurityPolicy `json:"items" protobuf:"bytes,2,rep,name=items"`
}
