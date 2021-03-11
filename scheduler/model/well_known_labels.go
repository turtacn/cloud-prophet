//
package model

const (
	JDLabelMachine       = "jdcloud/machine"
	JDLabelServiceType   = "jdcloud/service-type"
	JDLabelServiceCode   = "jdcloud/service-code"
	JDLabelAvailableZone = "jdcloud/az"
	JDLabelRegion        = "jdcloud/region"
	JDLabelPool          = "jdcloud/pool"
	JDLabelCluster       = "jdcloud/cluster"
)

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

	LabelWindowsBuild = "node.free.io/windows-build"

	LabelNamespaceSuffixKubelet = "kubelet.free.io"
	LabelNamespaceSuffixNode    = "node.free.io"

	LabelNamespaceNodeRestriction = "node-restriction.free.io"
)

func (n *Node) GetLabels() map[string]string {
	return n.Labels
}
