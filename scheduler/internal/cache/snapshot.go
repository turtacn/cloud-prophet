package cache

import (
	"fmt"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

type Snapshot struct {
	nodeInfoMap                      map[string]*framework.NodeInfo
	nodeInfoList                     []*framework.NodeInfo
	havePodsWithAffinityNodeInfoList []*framework.NodeInfo
	generation                       int64
}

var _ framework.SharedLister = &Snapshot{}

func NewEmptySnapshot() *Snapshot {
	return &Snapshot{
		nodeInfoMap: make(map[string]*framework.NodeInfo),
	}
}

func NewSnapshot(pods []*v1.Pod, nodes []*v1.Node) *Snapshot {
	nodeInfoMap := createNodeInfoMap(pods, nodes)
	nodeInfoList := make([]*framework.NodeInfo, 0, len(nodeInfoMap))
	havePodsWithAffinityNodeInfoList := make([]*framework.NodeInfo, 0, len(nodeInfoMap))
	for _, v := range nodeInfoMap {
		nodeInfoList = append(nodeInfoList, v)
		if len(v.PodsWithAffinity) > 0 {
			havePodsWithAffinityNodeInfoList = append(havePodsWithAffinityNodeInfoList, v)
		}
	}

	s := NewEmptySnapshot()
	s.nodeInfoMap = nodeInfoMap
	s.nodeInfoList = nodeInfoList
	s.havePodsWithAffinityNodeInfoList = havePodsWithAffinityNodeInfoList

	return s
}

func createNodeInfoMap(pods []*v1.Pod, nodes []*v1.Node) map[string]*framework.NodeInfo {
	nodeNameToInfo := make(map[string]*framework.NodeInfo)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if _, ok := nodeNameToInfo[nodeName]; !ok {
			nodeNameToInfo[nodeName] = framework.NewNodeInfo()
		}
		nodeNameToInfo[nodeName].AddPod(pod)
	}
	imageExistenceMap := createImageExistenceMap(nodes)

	for _, node := range nodes {
		if _, ok := nodeNameToInfo[node.Name]; !ok {
			nodeNameToInfo[node.Name] = framework.NewNodeInfo()
		}
		nodeInfo := nodeNameToInfo[node.Name]
		nodeInfo.SetNode(node)
		nodeInfo.ImageStates = getNodeImageStates(node, imageExistenceMap)
	}
	return nodeNameToInfo
}

func getNodeImageStates(node *v1.Node, imageExistenceMap map[string]sets.String) map[string]*framework.ImageStateSummary {
	imageStates := make(map[string]*framework.ImageStateSummary)

	return imageStates
}

func createImageExistenceMap(nodes []*v1.Node) map[string]sets.String {
	imageExistenceMap := make(map[string]sets.String)
	return imageExistenceMap
}

func (s *Snapshot) NodeInfos() framework.NodeInfoLister {
	return s
}

func (s *Snapshot) NumNodes() int {
	return len(s.nodeInfoList)
}

func (s *Snapshot) List() ([]*framework.NodeInfo, error) {
	return s.nodeInfoList, nil
}

func (s *Snapshot) HavePodsWithAffinityList() ([]*framework.NodeInfo, error) {
	return s.havePodsWithAffinityNodeInfoList, nil
}

func (s *Snapshot) Get(nodeName string) (*framework.NodeInfo, error) {
	if v, ok := s.nodeInfoMap[nodeName]; ok && v.Node() != nil {
		return v, nil
	}
	return nil, fmt.Errorf("nodeinfo not found for node name %q", nodeName)
}
