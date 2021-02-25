package k8s

type NodeUpdater interface {
	UpdateNode(name string, info *NodeInfo) error
}
