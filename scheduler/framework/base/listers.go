package base

type NodeInfoLister interface {
	List() ([]*NodeInfo, error)
	HavePodsWithAffinityList() ([]*NodeInfo, error)
	Get(nodeName string) (*NodeInfo, error)
}

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
