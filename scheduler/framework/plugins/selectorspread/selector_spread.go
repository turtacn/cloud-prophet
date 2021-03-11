// 在给定标签下的默认 spread 策略，分级分享权重
package selectorspread

import (
	"context"
	"fmt"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/helper"
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	utilnode "github.com/turtacn/cloud-prophet/scheduler/helper/node"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

type SelectorSpread struct {
	sharedLister framework.SharedLister
}

var _ framework.PreScorePlugin = &SelectorSpread{}
var _ framework.ScorePlugin = &SelectorSpread{}

const (
	Name             = "SelectorSpread"
	preScoreStateKey = "PreScore" + Name

	zoneWeighting float64 = 2.0 / 3.0
)

func (pl *SelectorSpread) Name() string {
	return Name
}

type preScoreState struct {
	selector labels.Selector
}

func (s *preScoreState) Clone() framework.StateData {
	return s
}

func skipSelectorSpread(pod *v1.Pod) bool {
	return len(pod.Spec.TopologySpreadConstraints) != 0
}

func (pl *SelectorSpread) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	if skipSelectorSpread(pod) {
		return 0, nil
	}

	c, err := state.Read(preScoreStateKey)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Error reading %q from cycleState: %v", preScoreStateKey, err))
	}

	s, ok := c.(*preScoreState)
	if !ok {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("%+v convert to tainttoleration.preScoreState error", c))
	}

	nodeInfo, err := pl.sharedLister.NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	count := countMatchingPods(pod.Namespace, s.selector, nodeInfo)
	return int64(count), nil
}

func (pl *SelectorSpread) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	if skipSelectorSpread(pod) {
		return nil
	}

	countsByZone := make(map[string]int64, 10)
	maxCountByZone := int64(0)
	maxCountByNodeName := int64(0)

	for i := range scores {
		if scores[i].Score > maxCountByNodeName {
			maxCountByNodeName = scores[i].Score
		}
		nodeInfo, err := pl.sharedLister.NodeInfos().Get(scores[i].Name)
		if err != nil {
			return framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", scores[i].Name, err))
		}
		zoneID := utilnode.GetZoneKey(nodeInfo.Node())
		if zoneID == "" {
			continue
		}
		countsByZone[zoneID] += scores[i].Score
	}

	for zoneID := range countsByZone {
		if countsByZone[zoneID] > maxCountByZone {
			maxCountByZone = countsByZone[zoneID]
		}
	}

	haveZones := len(countsByZone) != 0

	maxCountByNodeNameFloat64 := float64(maxCountByNodeName)
	maxCountByZoneFloat64 := float64(maxCountByZone)
	MaxNodeScoreFloat64 := float64(framework.MaxNodeScore)

	for i := range scores {
		fScore := MaxNodeScoreFloat64
		if maxCountByNodeName > 0 {
			fScore = MaxNodeScoreFloat64 * (float64(maxCountByNodeName-scores[i].Score) / maxCountByNodeNameFloat64)
		}
		if haveZones {
			nodeInfo, err := pl.sharedLister.NodeInfos().Get(scores[i].Name)
			if err != nil {
				return framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", scores[i].Name, err))
			}

			zoneID := utilnode.GetZoneKey(nodeInfo.Node())
			if zoneID != "" {
				zoneScore := MaxNodeScoreFloat64
				if maxCountByZone > 0 {
					zoneScore = MaxNodeScoreFloat64 * (float64(maxCountByZone-countsByZone[zoneID]) / maxCountByZoneFloat64)
				}
				fScore = (fScore * (1.0 - zoneWeighting)) + (zoneWeighting * zoneScore)
			}
		}
		scores[i].Score = int64(fScore)
	}
	return nil
}

func (pl *SelectorSpread) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

func (pl *SelectorSpread) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if skipSelectorSpread(pod) {
		return nil
	}
	var selector labels.Selector
	selector = helper.DefaultSelector(
		pod,
	)
	state := &preScoreState{
		selector: selector,
	}
	cycleState.Write(preScoreStateKey, state)
	return nil
}

func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	sharedLister := handle.SnapshotSharedLister()
	if sharedLister == nil {
		return nil, fmt.Errorf("SnapshotSharedLister is nil")
	}
	return &SelectorSpread{
		sharedLister: sharedLister,
	}, nil
}

func countMatchingPods(namespace string, selector labels.Selector, nodeInfo *framework.NodeInfo) int {
	if len(nodeInfo.Pods) == 0 || selector.Empty() {
		return 0
	}
	count := 0
	for _, p := range nodeInfo.Pods {
		if namespace == p.Pod.Namespace && p.Pod.DeletionTimestamp == nil {
			if selector.Matches(labels.Set(p.Pod.Labels)) {
				count++
			}
		}
	}
	return count
}
