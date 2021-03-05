//
//
package base

// NodeInfoLister interface represents anything that can list/get NodeInfo objects from node name.
type NodeInfoLister interface {
	// Returns the list of NodeInfos.
	List() ([]*NodeInfo, error)
	// Returns the list of NodeInfos of nodes with pods with affinity terms.
	HavePodsWithAffinityList() ([]*NodeInfo, error)
	// Returns the NodeInfo of the given node name.
	Get(nodeName string) (*NodeInfo, error)
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	NodeInfos() NodeInfoLister
}

type SharedPodsLister interface {
	PodInfos() PodInfoLister
}

type PodInfoLister interface {
	List(string) ([]*PodInfo, error) // host has pods
	Get(podName string) (*PodInfo, error)
}
