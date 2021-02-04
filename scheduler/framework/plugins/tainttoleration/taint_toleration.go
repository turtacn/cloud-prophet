package tainttoleration

import (
	"context"
	"fmt"

	pluginhelper "github.com/turtacn/cloud-prophet/scheduler/framework/plugins/helper"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/v1alpha1"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

// TaintToleration is a plugin that checks if a pod tolerates a node's taints.
type TaintToleration struct {
	handle framework.FrameworkHandle
}

var _ framework.FilterPlugin = &TaintToleration{}
var _ framework.PreScorePlugin = &TaintToleration{}
var _ framework.ScorePlugin = &TaintToleration{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "TaintToleration"
	// preScoreStateKey is the key in CycleState to TaintToleration pre-computed data for Scoring.
	preScoreStateKey = "PreScore" + Name
	// ErrReasonNotMatch is the Filter reason status when not matching.
	ErrReasonNotMatch = "node(s) had taints that the pod didn't tolerate"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *TaintToleration) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *TaintToleration) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return framework.NewStatus(framework.Error, "invalid nodeInfo")
	}

	filterPredicate := func(t *v1.Taint) bool {
		// PodToleratesNodeTaints is only interested in NoSchedule and NoExecute taints.
		return t.Effect == v1.TaintEffectNoSchedule || t.Effect == v1.TaintEffectNoExecute
	}

	taint := &v1.Taint{}
	isUntolerated := filterPredicate(taint)
	if !isUntolerated {
		return nil
	}

	errReason := fmt.Sprintf("node(s) had taint {%s, %s}, that the pod didn't tolerate",
		taint.Key, taint.Value)
	return framework.NewStatus(framework.UnschedulableAndUnresolvable, errReason)
}

// preScoreState computed at PreScore and used at Score.
type preScoreState struct {
	tolerationsPreferNoSchedule []v1.Toleration
}

// Clone implements the mandatory Clone interface. We don't really copy the data since
// there is no need for that.
func (s *preScoreState) Clone() framework.StateData {
	return s
}

// getAllTolerationEffectPreferNoSchedule gets the list of all Tolerations with Effect PreferNoSchedule or with no effect.
func getAllTolerationPreferNoSchedule(tolerations []v1.Toleration) (tolerationList []v1.Toleration) {
	for _, toleration := range tolerations {
		// Empty effect means all effects which includes PreferNoSchedule, so we need to collect it as well.
		if len(toleration.Effect) == 0 || toleration.Effect == v1.TaintEffectPreferNoSchedule {
			tolerationList = append(tolerationList, toleration)
		}
	}
	return
}

// PreScore builds and writes cycle state used by Score and NormalizeScore.
func (pl *TaintToleration) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if len(nodes) == 0 {
		return nil
	}
	tolerationsPreferNoSchedule := getAllTolerationPreferNoSchedule(pod.Spec.Tolerations)
	state := &preScoreState{
		tolerationsPreferNoSchedule: tolerationsPreferNoSchedule,
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
		return nil, fmt.Errorf("%+v convert to tainttoleration.preScoreState error", c)
	}
	return s, nil
}

// CountIntolerableTaintsPreferNoSchedule gives the count of intolerable taints of a pod with effect PreferNoSchedule
func countIntolerableTaintsPreferNoSchedule(taints []v1.Taint, tolerations []v1.Toleration) (intolerableTaints int) {
	for _, taint := range taints {
		// check only on taints that have effect PreferNoSchedule
		if taint.Effect != v1.TaintEffectPreferNoSchedule {
			continue
		}

		intolerableTaints++
	}
	return
}

// Score invoked at the Score extension point.
func (pl *TaintToleration) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	node := nodeInfo.Node()

	s, err := getPreScoreState(state)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	score := int64(countIntolerableTaintsPreferNoSchedule(node.Spec.Taints, s.tolerationsPreferNoSchedule))
	return score, nil
}

// NormalizeScore invoked after scoring all nodes.
func (pl *TaintToleration) NormalizeScore(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return pluginhelper.DefaultNormalizeScore(framework.MaxNodeScore, true, scores)
}

// ScoreExtensions of the Score plugin.
func (pl *TaintToleration) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &TaintToleration{handle: h}, nil
}
