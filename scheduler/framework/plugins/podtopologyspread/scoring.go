package podtopologyspread

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	pluginhelper "github.com/turtacn/cloud-prophet/scheduler/framework/plugins/helper"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	"github.com/turtacn/cloud-prophet/scheduler/internal/parallelize"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

const preScoreStateKey = "PreScore" + Name

type preScoreState struct {
	Constraints               []topologySpreadConstraint
	IgnoredNodes              sets.String
	TopologyPairToPodCounts   map[topologyPair]*int64
	TopologyNormalizingWeight []float64
}

func (s *preScoreState) Clone() framework.StateData {
	return s
}

func (pl *PodTopologySpread) initPreScoreState(s *preScoreState, pod *v1.Pod, filteredNodes []*v1.Node) error {
	var err error
	if len(pod.Spec.TopologySpreadConstraints) > 0 {
		s.Constraints, err = filterTopologySpreadConstraints(pod.Spec.TopologySpreadConstraints, v1.ScheduleAnyway)
		if err != nil {
			return fmt.Errorf("obtaining pod's soft topology spread constraints: %v", err)
		}
	} else {
		s.Constraints, err = pl.defaultConstraints(pod, v1.ScheduleAnyway)
		if err != nil {
			return fmt.Errorf("setting default soft topology spread constraints: %v", err)
		}
	}
	if len(s.Constraints) == 0 {
		return nil
	}
	topoSize := make([]int, len(s.Constraints))
	for _, node := range filteredNodes {
		if !nodeLabelsMatchSpreadConstraints(node.Labels, s.Constraints) {
			s.IgnoredNodes.Insert(node.Name)
			continue
		}
		for i, constraint := range s.Constraints {
			if constraint.TopologyKey == v1.LabelHostname {
				continue
			}
			pair := topologyPair{key: constraint.TopologyKey, value: node.Labels[constraint.TopologyKey]}
			if s.TopologyPairToPodCounts[pair] == nil {
				s.TopologyPairToPodCounts[pair] = new(int64)
				topoSize[i]++
			}
		}
	}

	s.TopologyNormalizingWeight = make([]float64, len(s.Constraints))
	for i, c := range s.Constraints {
		sz := topoSize[i]
		if c.TopologyKey == v1.LabelHostname {
			sz = len(filteredNodes) - len(s.IgnoredNodes)
		}
		s.TopologyNormalizingWeight[i] = topologyNormalizingWeight(sz)
	}
	return nil
}

func (pl *PodTopologySpread) PreScore(
	ctx context.Context,
	cycleState *framework.CycleState,
	pod *v1.Pod,
	filteredNodes []*v1.Node,
) *framework.Status {
	allNodes, err := pl.sharedLister.NodeInfos().List()
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("error when getting all nodes: %v", err))
	}

	if len(filteredNodes) == 0 || len(allNodes) == 0 {
		return nil
	}

	state := &preScoreState{
		IgnoredNodes:            sets.NewString(),
		TopologyPairToPodCounts: make(map[topologyPair]*int64),
	}
	err = pl.initPreScoreState(state, pod, filteredNodes)
	if err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("error when calculating preScoreState: %v", err))
	}

	if len(state.Constraints) == 0 {
		cycleState.Write(preScoreStateKey, state)
		return nil
	}

	processAllNode := func(i int) {
		nodeInfo := allNodes[i]
		node := nodeInfo.Node()
		if node == nil {
			return
		}
		if !pluginhelper.PodMatchesNodeSelectorAndAffinityTerms(pod, node) ||
			!nodeLabelsMatchSpreadConstraints(node.Labels, state.Constraints) {
			return
		}

		for _, c := range state.Constraints {
			pair := topologyPair{key: c.TopologyKey, value: node.Labels[c.TopologyKey]}
			tpCount := state.TopologyPairToPodCounts[pair]
			if tpCount == nil {
				continue
			}
			count := countPodsMatchSelector(nodeInfo.Pods, c.Selector, pod.Namespace)
			atomic.AddInt64(tpCount, int64(count))
		}
	}
	parallelize.Until(ctx, len(allNodes), processAllNode)

	cycleState.Write(preScoreStateKey, state)
	return nil
}

func (pl *PodTopologySpread) Score(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.sharedLister.NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v, node is nil: %v", nodeName, err, nodeInfo.Node() == nil))
	}

	node := nodeInfo.Node()
	s, err := getPreScoreState(cycleState)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	if s.IgnoredNodes.Has(node.Name) {
		return 0, nil
	}

	var score float64
	for i, c := range s.Constraints {
		if tpVal, ok := node.Labels[c.TopologyKey]; ok {
			var cnt int64
			if c.TopologyKey == v1.LabelHostname {
				cnt = int64(countPodsMatchSelector(nodeInfo.Pods, c.Selector, pod.Namespace))
			} else {
				pair := topologyPair{key: c.TopologyKey, value: tpVal}
				cnt = *s.TopologyPairToPodCounts[pair]
			}
			score += scoreForCount(cnt, c.MaxSkew, s.TopologyNormalizingWeight[i])
		}
	}
	return int64(score), nil
}

func (pl *PodTopologySpread) NormalizeScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	s, err := getPreScoreState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	if s == nil {
		return nil
	}

	var minScore int64 = math.MaxInt64
	var maxScore int64
	for _, score := range scores {
		if s.IgnoredNodes.Has(score.Name) {
			continue
		}
		if score.Score < minScore {
			minScore = score.Score
		}
		if score.Score > maxScore {
			maxScore = score.Score
		}
	}

	for i := range scores {
		nodeInfo, err := pl.sharedLister.NodeInfos().Get(scores[i].Name)
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		node := nodeInfo.Node()

		if s.IgnoredNodes.Has(node.Name) {
			scores[i].Score = 0
			continue
		}

		if maxScore == 0 {
			scores[i].Score = framework.MaxNodeScore
			continue
		}

		s := scores[i].Score
		scores[i].Score = framework.MaxNodeScore * (maxScore + minScore - s) / maxScore
	}
	return nil
}

func (pl *PodTopologySpread) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func getPreScoreState(cycleState *framework.CycleState) (*preScoreState, error) {
	c, err := cycleState.Read(preScoreStateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preScoreStateKey, err)
	}

	s, ok := c.(*preScoreState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to podtopologyspread.preScoreState error", c)
	}
	return s, nil
}

func topologyNormalizingWeight(size int) float64 {
	return math.Log(float64(size + 2))
}

func scoreForCount(cnt int64, maxSkew int32, tpWeight float64) float64 {
	return float64(cnt)*tpWeight + float64(maxSkew-1)
}
