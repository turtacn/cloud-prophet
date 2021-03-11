//
package interpodaffinity

import (
	"context"
	"fmt"
	"sync/atomic"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/internal/parallelize"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	schedutil "github.com/turtacn/cloud-prophet/scheduler/util"
)

const preScoreStateKey = "PreScore" + Name

type scoreMap map[string]map[string]int64

type preScoreState struct {
	topologyScore scoreMap
	podInfo       *framework.PodInfo
}

func (s *preScoreState) Clone() framework.StateData {
	return s
}

func (m scoreMap) processTerm(
	term *framework.WeightedAffinityTerm,
	podToCheck *v1.Pod,
	fixedNode *v1.Node,
	multiplier int,
) {
	if len(fixedNode.Labels) == 0 {
		return
	}

	match := schedutil.PodMatchesTermsNamespaceAndSelector(podToCheck, term.Namespaces, term.Selector)
	tpValue, tpValueExist := fixedNode.Labels[term.TopologyKey]
	if match && tpValueExist {
		if m[term.TopologyKey] == nil {
			m[term.TopologyKey] = make(map[string]int64)
		}
		m[term.TopologyKey][tpValue] += int64(term.Weight * int32(multiplier))
	}
	return
}

func (m scoreMap) processTerms(terms []framework.WeightedAffinityTerm, podToCheck *v1.Pod, fixedNode *v1.Node, multiplier int) {
	for _, term := range terms {
		m.processTerm(&term, podToCheck, fixedNode, multiplier)
	}
}

func (m scoreMap) append(other scoreMap) {
	for topology, oScores := range other {
		scores := m[topology]
		if scores == nil {
			m[topology] = oScores
			continue
		}
		for k, v := range oScores {
			scores[k] += v
		}
	}
}

func (pl *InterPodAffinity) processExistingPod(
	state *preScoreState,
	existingPod *framework.PodInfo,
	existingPodNodeInfo *framework.NodeInfo,
	incomingPod *v1.Pod,
	topoScore scoreMap,
) {
	existingPodNode := existingPodNodeInfo.Node()

	topoScore.processTerms(state.podInfo.PreferredAffinityTerms, existingPod.Pod, existingPodNode, 1)

	topoScore.processTerms(state.podInfo.PreferredAntiAffinityTerms, existingPod.Pod, existingPodNode, -1)

	if pl.args.HardPodAffinityWeight > 0 {
		for _, term := range existingPod.RequiredAffinityTerms {
			t := framework.WeightedAffinityTerm{AffinityTerm: term, Weight: pl.args.HardPodAffinityWeight}
			topoScore.processTerm(&t, incomingPod, existingPodNode, 1)
		}
	}

	topoScore.processTerms(existingPod.PreferredAffinityTerms, incomingPod, existingPodNode, 1)

	topoScore.processTerms(existingPod.PreferredAntiAffinityTerms, incomingPod, existingPodNode, -1)
}

func (pl *InterPodAffinity) PreScore(
	pCtx context.Context,
	cycleState *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) *framework.Status {
	if len(nodes) == 0 {
		return nil
	}

	if pl.sharedLister == nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("InterPodAffinity PreScore with empty shared lister found"))
	}

	affinity := pod.Spec.Affinity
	hasAffinityConstraints := affinity != nil && affinity.PodAffinity != nil
	hasAntiAffinityConstraints := affinity != nil && affinity.PodAntiAffinity != nil

	var allNodes []*framework.NodeInfo
	var err error
	if hasAffinityConstraints || hasAntiAffinityConstraints {
		allNodes, err = pl.sharedLister.NodeInfos().List()
		if err != nil {
			framework.NewStatus(framework.Error, fmt.Sprintf("get all nodes from shared lister error, err: %v", err))
		}
	} else {
		allNodes, err = pl.sharedLister.NodeInfos().HavePodsWithAffinityList()
		if err != nil {
			framework.NewStatus(framework.Error, fmt.Sprintf("get pods with affinity list error, err: %v", err))
		}
	}

	podInfo := framework.NewPodInfo(pod)
	if podInfo.ParseError != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("parsing pod: %+v", podInfo.ParseError))
	}

	state := &preScoreState{
		topologyScore: make(map[string]map[string]int64),
		podInfo:       podInfo,
	}

	topoScores := make([]scoreMap, len(allNodes))
	index := int32(-1)
	processNode := func(i int) {
		nodeInfo := allNodes[i]
		if nodeInfo.Node() == nil {
			return
		}
		podsToProcess := nodeInfo.PodsWithAffinity
		if hasAffinityConstraints || hasAntiAffinityConstraints {
			podsToProcess = nodeInfo.Pods
		}

		topoScore := make(scoreMap)
		for _, existingPod := range podsToProcess {
			pl.processExistingPod(state, existingPod, nodeInfo, pod, topoScore)
		}
		if len(topoScore) > 0 {
			topoScores[atomic.AddInt32(&index, 1)] = topoScore
		}
	}
	parallelize.Until(context.Background(), len(allNodes), processNode)

	for i := 0; i <= int(index); i++ {
		state.topologyScore.append(topoScores[i])
	}

	cycleState.Write(preScoreStateKey, state)
	return nil
}

func getPreScoreState(cycleState *framework.CycleState) (*preScoreState, error) {
	c, err := cycleState.Read(preScoreStateKey)
	if err != nil {
		return nil, fmt.Errorf("Error reading %q from cycleState: %v", preScoreStateKey, err)
	}

	s, ok := c.(*preScoreState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to interpodaffinity.preScoreState error", c)
	}
	return s, nil
}

func (pl *InterPodAffinity) Score(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.sharedLister.NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v, node is nil: %v", nodeName, err, nodeInfo.Node() == nil))
	}
	node := nodeInfo.Node()

	s, err := getPreScoreState(cycleState)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}
	var score int64
	for tpKey, tpValues := range s.topologyScore {
		if v, exist := node.Labels[tpKey]; exist {
			score += tpValues[v]
		}
	}

	return score, nil
}

func (pl *InterPodAffinity) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	s, err := getPreScoreState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	if len(s.topologyScore) == 0 {
		return nil
	}

	var maxCount, minCount int64
	for i := range scores {
		score := scores[i].Score
		if score > maxCount {
			maxCount = score
		}
		if score < minCount {
			minCount = score
		}
	}

	maxMinDiff := maxCount - minCount
	for i := range scores {
		fScore := float64(0)
		if maxMinDiff > 0 {
			fScore = float64(framework.MaxNodeScore) * (float64(scores[i].Score-minCount) / float64(maxMinDiff))
		}

		scores[i].Score = int64(fScore)
	}

	return nil
}

func (pl *InterPodAffinity) ScoreExtensions() framework.ScoreExtensions {
	return pl
}
