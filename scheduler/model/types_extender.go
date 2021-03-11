package model

const (
	MinExtenderPriority int64 = 0

	MaxExtenderPriority int64 = 10
)

type ExtenderPreemptionResult struct {
	NodeNameToMetaVictims map[string]*MetaVictims
}

type ExtenderPreemptionArgs struct {
	Pod                   *Pod
	NodeNameToVictims     map[string]*Victims
	NodeNameToMetaVictims map[string]*MetaVictims
}

type Victims struct {
	Pods             []*Pod
	NumPDBViolations int64
}

type MetaPod struct {
	UID string
}

type MetaVictims struct {
	Pods             []*MetaPod
	NumPDBViolations int64
}

type ExtenderArgs struct {
	Pod       *Pod
	Nodes     *NodeList
	NodeNames *[]string
}

type FailedNodesMap map[string]string

type ExtenderFilterResult struct {
	Nodes       *NodeList
	NodeNames   *[]string
	FailedNodes FailedNodesMap
	Error       string
}

type ExtenderBindingArgs struct {
	PodName      string
	PodNamespace string
	PodUID       string
	Node         string
}

type ExtenderBindingResult struct {
	Error string
}

type HostPriority struct {
	Host  string
	Score int64
}

type HostPriorityList []HostPriority
