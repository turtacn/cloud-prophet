package nodeunschedulable

import (
	"context"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/v1alpha1"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

// NodeUnschedulable is a plugin that priorities nodes according to the node annotation
// "scheduler.alpha.kubernetes.io/preferAvoidPods".
type NodeUnschedulable struct {
}

// 编译时接口检查
var _ framework.FilterPlugin = &NodeUnschedulable{}

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "NodeUnschedulable"

const (
	// ErrReasonUnknownCondition is used for NodeUnknownCondition predicate error.
	ErrReasonUnknownCondition = "node(s) had unknown conditions"
	// ErrReasonUnschedulable is used for NodeUnschedulable predicate error.
	ErrReasonUnschedulable = "node(s) were unschedulable"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeUnschedulable) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
// 目前不支持对 Unschedulable (jvirt disable) 节点继续调度的业务逻辑
func (pl *NodeUnschedulable) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonUnknownCondition)
	}
	// If pod tolerate unschedulable taint, it's also tolerate `node.Spec.Unschedulable`.
	podToleratesUnschedulable := false
	//	v1helper.TolerationsTolerateTaint(pod.Spec.Tolerations, &v1.Taint{
	//	Key:    v1.TaintNodeUnschedulable,
	//	Effect: v1.TaintEffectNoSchedule,
	//})

	if nodeInfo.Node().Spec.Unschedulable && !podToleratesUnschedulable {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonUnschedulable)
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodeUnschedulable{}, nil
}
