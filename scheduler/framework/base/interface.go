package base

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

type NodeScoreList []NodeScore

type NodeScore struct {
	Name  string
	Score int64
}

type PluginToNodeScores map[string]NodeScoreList

type NodeToStatusMap map[string]*Status

type Code int

const (
	Success Code = iota
	Error
	Unschedulable
	UnschedulableAndUnresolvable
	Wait
	Skip
)

var codes = []string{"Success", "Error", "Unschedulable", "UnschedulableAndUnresolvable", "Wait", "Skip"}

func (c Code) String() string {
	return codes[c]
}

const (
	MaxNodeScore int64 = 100

	MinNodeScore int64 = 0

	MaxTotalScore int64 = math.MaxInt64
)

type Status struct {
	code    Code
	reasons []string
}

func (s *Status) Code() Code {
	if s == nil {
		return Success
	}
	return s.code
}

func (s *Status) Message() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.reasons, ", ")
}

func (s *Status) Reasons() []string {
	return s.reasons
}

func (s *Status) AppendReason(reason string) {
	s.reasons = append(s.reasons, reason)
}

func (s *Status) IsSuccess() bool {
	return s.Code() == Success
}

func (s *Status) IsUnschedulable() bool {
	code := s.Code()
	return code == Unschedulable || code == UnschedulableAndUnresolvable
}

func (s *Status) AsError() error {
	if s.IsSuccess() {
		return nil
	}
	return errors.New(s.Message())
}

func NewStatus(code Code, reasons ...string) *Status {
	return &Status{
		code:    code,
		reasons: reasons,
	}
}

type PluginToStatus map[string]*Status

func (p PluginToStatus) Merge() *Status {
	if len(p) == 0 {
		return nil
	}

	finalStatus := NewStatus(Success)
	var hasError, hasUnschedulableAndUnresolvable, hasUnschedulable bool
	for _, s := range p {
		if s.Code() == Error {
			hasError = true
		} else if s.Code() == UnschedulableAndUnresolvable {
			hasUnschedulableAndUnresolvable = true
		} else if s.Code() == Unschedulable {
			hasUnschedulable = true
		}
		finalStatus.code = s.Code()
		for _, r := range s.reasons {
			finalStatus.AppendReason(r)
		}
	}

	if hasError {
		finalStatus.code = Error
	} else if hasUnschedulableAndUnresolvable {
		finalStatus.code = UnschedulableAndUnresolvable
	} else if hasUnschedulable {
		finalStatus.code = Unschedulable
	}
	return finalStatus
}

type WaitingPod interface {
	GetPod() *v1.Pod
	GetPendingPlugins() []string
	Allow(pluginName string)
	Reject(msg string)
}

type Plugin interface {
	Name() string
}

type LessFunc func(podInfo1, podInfo2 *QueuedPodInfo) bool

type QueueSortPlugin interface {
	Plugin
	Less(*QueuedPodInfo, *QueuedPodInfo) bool
}

type PreFilterExtensions interface {
	AddPod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *NodeInfo) *Status
	RemovePod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *NodeInfo) *Status
}

type PreFilterPlugin interface {
	Plugin
	PreFilter(ctx context.Context, state *CycleState, p *v1.Pod) *Status
	PreFilterExtensions() PreFilterExtensions
}

type FilterPlugin interface {
	Plugin
	Filter(ctx context.Context, state *CycleState, pod *v1.Pod, nodeInfo *NodeInfo) *Status
}

type PostFilterPlugin interface {
	Plugin
	PostFilter(ctx context.Context, state *CycleState, pod *v1.Pod, filteredNodeStatusMap NodeToStatusMap) (*PostFilterResult, *Status)
}

type PreScorePlugin interface {
	Plugin
	PreScore(ctx context.Context, state *CycleState, pod *v1.Pod, nodes []*v1.Node) *Status
}

type ScoreExtensions interface {
	NormalizeScore(ctx context.Context, state *CycleState, p *v1.Pod, scores NodeScoreList) *Status
}

type ScorePlugin interface {
	Plugin
	Score(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string) (int64, *Status)

	ScoreExtensions() ScoreExtensions
}

type ReservePlugin interface {
	Plugin
	Reserve(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string) *Status
	Unreserve(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string)
}

type PreBindPlugin interface {
	Plugin
	PreBind(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string) *Status
}

type PostBindPlugin interface {
	Plugin
	PostBind(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string)
}

type PermitPlugin interface {
	Plugin
	Permit(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string) (*Status, time.Duration)
}

type BindPlugin interface {
	Plugin
	Bind(ctx context.Context, state *CycleState, p *v1.Pod, nodeName string) *Status
}

type Framework interface {
	FrameworkHandle
	QueueSortFunc() LessFunc

	RunPreFilterPlugins(ctx context.Context, state *CycleState, pod *v1.Pod) *Status

	RunFilterPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodeInfo *NodeInfo) PluginToStatus

	RunPostFilterPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, filteredNodeStatusMap NodeToStatusMap) (*PostFilterResult, *Status)

	RunPreFilterExtensionAddPod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *NodeInfo) *Status

	RunPreFilterExtensionRemovePod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *NodeInfo) *Status

	RunPreScorePlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodes []*v1.Node) *Status

	RunScorePlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodes []*v1.Node) (PluginToNodeScores, *Status)

	RunPreBindPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string) *Status

	RunPostBindPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string)

	RunReservePluginsReserve(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string) *Status

	RunReservePluginsUnreserve(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string)

	RunPermitPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string) *Status

	WaitOnPermit(ctx context.Context, pod *v1.Pod) *Status

	RunBindPlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodeName string) *Status

	HasFilterPlugins() bool

	HasPostFilterPlugins() bool

	HasScorePlugins() bool

	ListPlugins() map[string][]config.Plugin
}

type FrameworkHandle interface {
	SnapshotSharedLister() SharedLister

	IterateOverWaitingPods(callback func(WaitingPod))

	GetWaitingPod(uid string) WaitingPod

	RejectWaitingPod(uid string)

	ClientSet() ClientSet

	PreemptHandle() PreemptHandle
}

type PostFilterResult struct {
	NominatedNodeName string
}

type PreemptHandle interface {
	PodNominator
	PluginsRunner
}

type PodNominator interface {
	AddNominatedPod(pod *v1.Pod, nodeName string)
	DeleteNominatedPodIfExists(pod *v1.Pod)
	UpdateNominatedPod(oldPod, newPod *v1.Pod)
	NominatedPodsForNode(nodeName string) []*v1.Pod
}

type PluginsRunner interface {
	RunFilterPlugins(context.Context, *CycleState, *v1.Pod, *NodeInfo) PluginToStatus
	RunPreFilterExtensionAddPod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *NodeInfo) *Status
	RunPreFilterExtensionRemovePod(ctx context.Context, state *CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *NodeInfo) *Status
}
