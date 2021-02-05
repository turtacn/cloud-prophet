//
package model

// jvirt 标签扩展
const (
	JDLabelMachine       = "jdcloud/machine"
	JDLabelServiceType   = "jdcloud/service-type"
	JDLabelServiceCode   = "jdcloud/service-code"
	JDLabelAvailableZone = "jdcloud/az"
	JDLabelRegion        = "jdcloud/region"
	JDLabelPool          = "jdcloud/pool"
	JDLabelCluster       = "jdcloud/cluster"
)

// kubernetes 标签
const (
	LabelHostname = "free.io/hostname"

	LabelZoneFailureDomain       = "failure-domain.beta.free.io/zone"
	LabelZoneRegion              = "failure-domain.beta.free.io/region"
	LabelZoneFailureDomainStable = "topology.free.io/zone"
	LabelZoneRegionStable        = "topology.free.io/region"

	LabelInstanceType       = "beta.free.io/instance-type"
	LabelInstanceTypeStable = "node.free.io/instance-type"

	LabelOSStable   = "free.io/os"
	LabelArchStable = "free.io/arch"

	// LabelWindowsBuild is used on Windows nodes to specify the Windows build number starting with v1.17.0.
	// It's in the format MajorVersion.MinorVersion.BuildNumber (for ex: 10.0.17763)
	LabelWindowsBuild = "node.free.io/windows-build"

	// LabelNamespaceSuffixKubelet is an allowed label namespace suffix kubelets can self-set ([*.]kubelet.free.io/*)
	LabelNamespaceSuffixKubelet = "kubelet.free.io"
	// LabelNamespaceSuffixNode is an allowed label namespace suffix kubelets can self-set ([*.]node.free.io/*)
	LabelNamespaceSuffixNode = "node.free.io"

	// LabelNamespaceNodeRestriction is a forbidden label namespace that kubelets may not self-set when the NodeRestriction admission plugin is enabled
	LabelNamespaceNodeRestriction = "node-restriction.free.io"
)

func (n *Node) GetLabels() map[string]string {
	return n.Labels
}
