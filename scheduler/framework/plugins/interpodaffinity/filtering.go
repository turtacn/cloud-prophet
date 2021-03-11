package interpodaffinity

import (
	"context"
	"fmt"
	"sync/atomic"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/internal/parallelize"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	schedutil "github.com/turtacn/cloud-prophet/scheduler/util"
	"k8s.io/klog/v2"
)

const (
	preFilterStateKey = "PreFilter" + Name

	ErrReasonExistingAntiAffinityRulesNotMatch = "node(s) didn't satisfy existing pods anti-affinity rules"
	ErrReasonAffinityNotMatch                  = "node(s) didn't match pod affinity/anti-affinity"
	ErrReasonAffinityRulesNotMatch             = "node(s) didn't match pod affinity rules"
	ErrReasonAntiAffinityRulesNotMatch         = "node(s) didn't match pod anti-affinity rules"
)

type preFilterState struct {
	topologyToMatchedExistingAntiAffinityTerms topologyToMatchedTermCount
	topologyToMatchedAffinityTerms             topologyToMatchedTermCount
	topologyToMatchedAntiAffinityTerms         topologyToMatchedTermCount
	podInfo                                    *framework.PodInfo
}

func (s *preFilterState) Clone() framework.StateData {
	if s == nil {
		return nil
	}

	copy := preFilterState{}
	copy.topologyToMatchedAffinityTerms = s.topologyToMatchedAffinityTerms.clone()
	copy.topologyToMatchedAntiAffinityTerms = s.topologyToMatchedAntiAffinityTerms.clone()
	copy.topologyToMatchedExistingAntiAffinityTerms = s.topologyToMatchedExistingAntiAffinityTerms.clone()
	copy.podInfo = s.podInfo

	return &copy
}

func (s *preFilterState) updateWithPod(updatedPod *v1.Pod, node *v1.Node, multiplier int64) error {
	if s == nil {
		return nil
	}

	updatedPodInfo := framework.NewPodInfo(updatedPod)
	s.topologyToMatchedExistingAntiAffinityTerms.updateWithAntiAffinityTerms(s.podInfo.Pod, node, updatedPodInfo.RequiredAntiAffinityTerms, multiplier)

	s.topologyToMatchedAffinityTerms.updateWithAffinityTerms(updatedPod, node, s.podInfo.RequiredAffinityTerms, multiplier)
	s.topologyToMatchedAntiAffinityTerms.updateWithAntiAffinityTerms(updatedPod, node, s.podInfo.RequiredAntiAffinityTerms, multiplier)

	return nil
}

type topologyPair struct {
	key   string
	value string
}
type topologyToMatchedTermCount map[topologyPair]int64

func (m topologyToMatchedTermCount) append(toAppend topologyToMatchedTermCount) {
	for pair := range toAppend {
		m[pair] += toAppend[pair]
	}
}

func (m topologyToMatchedTermCount) clone() topologyToMatchedTermCount {
	copy := make(topologyToMatchedTermCount, len(m))
	copy.append(m)
	return copy
}

func (m topologyToMatchedTermCount) updateWithAffinityTerms(targetPod *v1.Pod, targetPodNode *v1.Node, affinityTerms []framework.AffinityTerm, value int64) {
	if podMatchesAllAffinityTerms(targetPod, affinityTerms) {
		for _, t := range affinityTerms {
			if topologyValue, ok := targetPodNode.Labels[t.TopologyKey]; ok {
				pair := topologyPair{key: t.TopologyKey, value: topologyValue}
				m[pair] += value
				if m[pair] == 0 {
					delete(m, pair)
				}
			}
		}
	}
}

func (m topologyToMatchedTermCount) updateWithAntiAffinityTerms(targetPod *v1.Pod, targetPodNode *v1.Node, antiAffinityTerms []framework.AffinityTerm, value int64) {
	for _, a := range antiAffinityTerms {
		if schedutil.PodMatchesTermsNamespaceAndSelector(targetPod, a.Namespaces, a.Selector) {
			if topologyValue, ok := targetPodNode.Labels[a.TopologyKey]; ok {
				pair := topologyPair{key: a.TopologyKey, value: topologyValue}
				m[pair] += value
				if m[pair] == 0 {
					delete(m, pair)
				}
			}
		}
	}
}

func podMatchesAllAffinityTerms(pod *v1.Pod, terms []framework.AffinityTerm) bool {
	if len(terms) == 0 {
		return false
	}
	for _, term := range terms {
		if !schedutil.PodMatchesTermsNamespaceAndSelector(pod, term.Namespaces, term.Selector) {
			return false
		}
	}
	return true
}

func getTPMapMatchingExistingAntiAffinity(pod *v1.Pod, allNodes []*framework.NodeInfo) topologyToMatchedTermCount {
	topoMaps := make([]topologyToMatchedTermCount, len(allNodes))
	index := int32(-1)
	processNode := func(i int) {
		nodeInfo := allNodes[i]
		node := nodeInfo.Node()
		if node == nil {
			klog.Error("node not found")
			return
		}
		topoMap := make(topologyToMatchedTermCount)
		for _, existingPod := range nodeInfo.PodsWithAffinity {
			topoMap.updateWithAntiAffinityTerms(pod, node, existingPod.RequiredAntiAffinityTerms, 1)
		}
		if len(topoMap) != 0 {
			topoMaps[atomic.AddInt32(&index, 1)] = topoMap
		}
	}
	parallelize.Until(context.Background(), len(allNodes), processNode)

	result := make(topologyToMatchedTermCount)
	for i := 0; i <= int(index); i++ {
		result.append(topoMaps[i])
	}

	return result
}

func getTPMapMatchingIncomingAffinityAntiAffinity(podInfo *framework.PodInfo, allNodes []*framework.NodeInfo) (topologyToMatchedTermCount, topologyToMatchedTermCount) {
	affinityCounts := make(topologyToMatchedTermCount)
	antiAffinityCounts := make(topologyToMatchedTermCount)
	if len(podInfo.RequiredAffinityTerms) == 0 && len(podInfo.RequiredAntiAffinityTerms) == 0 {
		return affinityCounts, antiAffinityCounts
	}

	affinityCountsList := make([]topologyToMatchedTermCount, len(allNodes))
	antiAffinityCountsList := make([]topologyToMatchedTermCount, len(allNodes))
	index := int32(-1)
	processNode := func(i int) {
		nodeInfo := allNodes[i]
		node := nodeInfo.Node()
		if node == nil {
			klog.Error("node not found")
			return
		}
		affinity := make(topologyToMatchedTermCount)
		antiAffinity := make(topologyToMatchedTermCount)
		for _, existingPod := range nodeInfo.Pods {
			affinity.updateWithAffinityTerms(existingPod.Pod, node, podInfo.RequiredAffinityTerms, 1)

			antiAffinity.updateWithAntiAffinityTerms(existingPod.Pod, node, podInfo.RequiredAntiAffinityTerms, 1)
		}

		if len(affinity) > 0 || len(antiAffinity) > 0 {
			k := atomic.AddInt32(&index, 1)
			affinityCountsList[k] = affinity
			antiAffinityCountsList[k] = antiAffinity
		}
	}
	parallelize.Until(context.Background(), len(allNodes), processNode)

	for i := 0; i <= int(index); i++ {
		affinityCounts.append(affinityCountsList[i])
		antiAffinityCounts.append(antiAffinityCountsList[i])
	}

	return affinityCounts, antiAffinityCounts
}

func (pl *InterPodAffinity) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	var allNodes []*framework.NodeInfo
	var havePodsWithAffinityNodes []*framework.NodeInfo
	var err error
	if allNodes, err = pl.sharedLister.NodeInfos().List(); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("failed to list NodeInfos: %v", err))
	}
	if havePodsWithAffinityNodes, err = pl.sharedLister.NodeInfos().HavePodsWithAffinityList(); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("failed to list NodeInfos with pods with affinity: %v", err))
	}

	podInfo := framework.NewPodInfo(pod)
	if podInfo.ParseError != nil {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("parsing pod: %+v", podInfo.ParseError))
	}

	existingPodAntiAffinityMap := getTPMapMatchingExistingAntiAffinity(pod, havePodsWithAffinityNodes)

	incomingPodAffinityMap, incomingPodAntiAffinityMap := getTPMapMatchingIncomingAffinityAntiAffinity(podInfo, allNodes)

	s := &preFilterState{
		topologyToMatchedAffinityTerms:             incomingPodAffinityMap,
		topologyToMatchedAntiAffinityTerms:         incomingPodAntiAffinityMap,
		topologyToMatchedExistingAntiAffinityTerms: existingPodAntiAffinityMap,
		podInfo: podInfo,
	}

	cycleState.Write(preFilterStateKey, s)
	return nil
}

func (pl *InterPodAffinity) PreFilterExtensions() framework.PreFilterExtensions {
	return pl
}

func (pl *InterPodAffinity) AddPod(ctx context.Context, cycleState *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	state, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	state.updateWithPod(podToAdd, nodeInfo.Node(), 1)
	return nil
}

func (pl *InterPodAffinity) RemovePod(ctx context.Context, cycleState *framework.CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	state, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	state.updateWithPod(podToRemove, nodeInfo.Node(), -1)
	return nil
}

func getPreFilterState(cycleState *framework.CycleState) (*preFilterState, error) {
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to interpodaffinity.state error", c)
	}
	return s, nil
}

func satisfyExistingPodsAntiAffinity(state *preFilterState, nodeInfo *framework.NodeInfo) bool {
	if len(state.topologyToMatchedExistingAntiAffinityTerms) > 0 {
		for topologyKey, topologyValue := range nodeInfo.Node().Labels {
			tp := topologyPair{key: topologyKey, value: topologyValue}
			if state.topologyToMatchedExistingAntiAffinityTerms[tp] > 0 {
				return false
			}
		}
	}
	return true
}

func satisfyPodAntiAffinity(state *preFilterState, nodeInfo *framework.NodeInfo) bool {
	for _, term := range state.podInfo.RequiredAntiAffinityTerms {
		if topologyValue, ok := nodeInfo.Node().Labels[term.TopologyKey]; ok {
			tp := topologyPair{key: term.TopologyKey, value: topologyValue}
			if state.topologyToMatchedAntiAffinityTerms[tp] > 0 {
				return false
			}
		}
	}
	return true
}

func satisfyPodAffinity(state *preFilterState, nodeInfo *framework.NodeInfo) bool {
	podsExist := true
	for _, term := range state.podInfo.RequiredAffinityTerms {
		if topologyValue, ok := nodeInfo.Node().Labels[term.TopologyKey]; ok {
			tp := topologyPair{key: term.TopologyKey, value: topologyValue}
			if state.topologyToMatchedAffinityTerms[tp] <= 0 {
				podsExist = false
			}
		} else {
			return false
		}
	}

	if !podsExist {
		podInfo := state.podInfo
		if len(state.topologyToMatchedAffinityTerms) == 0 && podMatchesAllAffinityTerms(podInfo.Pod, podInfo.RequiredAffinityTerms) {
			return true
		}
		return false
	}
	return true
}

func (pl *InterPodAffinity) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo.Node() == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}

	state, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	if !satisfyPodAffinity(state, nodeInfo) {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonAffinityNotMatch, ErrReasonAffinityRulesNotMatch)
	}

	if !satisfyPodAntiAffinity(state, nodeInfo) {
		return framework.NewStatus(framework.Unschedulable, ErrReasonAffinityNotMatch, ErrReasonAntiAffinityRulesNotMatch)
	}

	if !satisfyExistingPodsAntiAffinity(state, nodeInfo) {
		return framework.NewStatus(framework.Unschedulable, ErrReasonAffinityNotMatch, ErrReasonExistingAntiAffinityRulesNotMatch)
	}

	return nil
}
