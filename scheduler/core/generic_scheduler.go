//
package core

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/klog/v2"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	podutil "github.com/turtacn/cloud-prophet/scheduler/helper/pod"
	internalcache "github.com/turtacn/cloud-prophet/scheduler/internal/cache"
	"github.com/turtacn/cloud-prophet/scheduler/internal/parallelize"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	utiltrace "k8s.io/utils/trace"
)

const (
	minFeasibleNodesToFind           = 100
	minFeasibleNodesPercentageToFind = 5
)

type FitError struct {
	Pod                   *v1.Pod
	NumAllNodes           int
	FilteredNodesStatuses framework.NodeToStatusMap
}

var ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule pods")

const (
	NoNodeAvailableMsg = "0/%v nodes are available"
)

func (f *FitError) Error() string {
	reasons := make(map[string]int)
	for _, status := range f.FilteredNodesStatuses {
		for _, reason := range status.Reasons() {
			reasons[reason]++
		}
	}

	sortReasonsHistogram := func() []string {
		var reasonStrings []string
		for k, v := range reasons {
			reasonStrings = append(reasonStrings, fmt.Sprintf("%v %v", v, k))
		}
		sort.Strings(reasonStrings)
		return reasonStrings
	}
	reasonMsg := fmt.Sprintf(NoNodeAvailableMsg+": %v.", f.NumAllNodes, strings.Join(sortReasonsHistogram(), ", "))
	return reasonMsg
}

type ScheduleAlgorithm interface {
	Schedule(context.Context, *profile.Profile, *framework.CycleState, *v1.Pod) (scheduleResult ScheduleResult, err error)
}

type ScheduleResult struct {
	SuggestedHost  string
	EvaluatedNodes int
	FeasibleNodes  int
}

type genericScheduler struct {
	cache                    internalcache.Cache
	nodeInfoSnapshot         *internalcache.Snapshot
	percentageOfNodesToScore int32
	nextStartNodeIndex       int
}

func (g *genericScheduler) snapshot() error {
	return g.cache.UpdateSnapshot(g.nodeInfoSnapshot)
}

func (g *genericScheduler) Schedule(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod) (result ScheduleResult, err error) {
	trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: pod.Namespace}, utiltrace.Field{Key: "name", Value: pod.Name})
	defer trace.LogIfLong(100 * time.Millisecond)

	if err := podPassesBasicChecks(pod); err != nil {
		return result, err
	}
	trace.Step("Basic checks done")

	if err := g.snapshot(); err != nil {
		klog.Errorf("snapshot from generic scheduler error %v", err)
		return result, err
	}
	trace.Step("Snapshotting scheduler cache and node infos done")

	if g.nodeInfoSnapshot.NumNodes() == 0 {
		klog.Errorf("node info snapshot error %v", ErrNoNodesAvailable)
		return result, ErrNoNodesAvailable
	}

	startPredicateEvalTime := time.Now()
	feasibleNodes, filteredNodesStatuses, err := g.findNodesThatFitPod(ctx, prof, state, pod)
	if err != nil {
		klog.Errorf("find nodes fit pod error %v", err)
		return result, err
	}
	trace.Step(fmt.Sprintf("Computing predicates done, time=%f", time.Now().Sub(startPredicateEvalTime).Seconds()))

	if len(feasibleNodes) == 0 {
		fitErr := &FitError{
			Pod:                   pod,
			NumAllNodes:           g.nodeInfoSnapshot.NumNodes(),
			FilteredNodesStatuses: filteredNodesStatuses,
		}
		klog.Errorf("feasible nodes not found error %v", fitErr)
		return result, fitErr
	}
	startPriorityEvalTime := time.Now()
	if len(feasibleNodes) == 1 {
		return ScheduleResult{
			SuggestedHost:  feasibleNodes[0].Name,
			EvaluatedNodes: 1 + len(filteredNodesStatuses),
			FeasibleNodes:  1,
		}, nil
	}

	priorityList, err := g.prioritizeNodes(ctx, prof, state, pod, feasibleNodes)
	if err != nil {
		return result, err
	}

	host, err := g.selectHost(priorityList)
	trace.Step(fmt.Sprintf("Prioritizing done, time=%f", time.Now().Sub(startPriorityEvalTime).Seconds()))

	return ScheduleResult{
		SuggestedHost:  host,
		EvaluatedNodes: len(feasibleNodes) + len(filteredNodesStatuses),
		FeasibleNodes:  len(feasibleNodes),
	}, err
}

func (g *genericScheduler) selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				selected = ns.Name
			}
		}
	}
	return selected, nil
}

func (g *genericScheduler) numFeasibleNodesToFind(numAllNodes int32) (numNodes int32) {
	if numAllNodes < minFeasibleNodesToFind || g.percentageOfNodesToScore >= 100 {
		return numAllNodes
	}

	adaptivePercentage := g.percentageOfNodesToScore
	if adaptivePercentage <= 0 {
		basePercentageOfNodesToScore := int32(50)
		adaptivePercentage = basePercentageOfNodesToScore - numAllNodes/125
		if adaptivePercentage < minFeasibleNodesPercentageToFind {
			adaptivePercentage = minFeasibleNodesPercentageToFind
		}
	}

	numNodes = numAllNodes * adaptivePercentage / 100
	if numNodes < minFeasibleNodesToFind {
		return minFeasibleNodesToFind
	}

	return numNodes
}

func (g *genericScheduler) findNodesThatFitPod(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod) ([]*v1.Node, framework.NodeToStatusMap, error) {
	filteredNodesStatuses := make(framework.NodeToStatusMap)

	s := prof.RunPreFilterPlugins(ctx, state, pod)
	if !s.IsSuccess() {
		if !s.IsUnschedulable() {
			klog.Errorf("unschedulable reason error %v", s.AsError())
			return nil, nil, s.AsError()
		}
		allNodes, err := g.nodeInfoSnapshot.NodeInfos().List()
		if err != nil {
			klog.Errorf("nodeinfo snapshot list nodes error %v", err)
			return nil, nil, err
		}
		for _, n := range allNodes {
			filteredNodesStatuses[n.Node().Name] = s
		}
		klog.Warningf("no fit nodes found and status %v", filteredNodesStatuses)
		return nil, filteredNodesStatuses, nil

	}
	feasibleNodes, err := g.findNodesThatPassFilters(ctx, prof, state, pod, filteredNodesStatuses)
	if err != nil {
		klog.Errorf("findNodesThatPassFilters error %v", err)
		return nil, nil, err
	}

	for k, v := range filteredNodesStatuses {
		klog.Infof("node %s status %v", k, v.Reasons())
		break
	}

	return feasibleNodes, filteredNodesStatuses, nil
}

func (g *genericScheduler) findNodesThatPassFilters(ctx context.Context, prof *profile.Profile, state *framework.CycleState, pod *v1.Pod, statuses framework.NodeToStatusMap) ([]*v1.Node, error) {
	allNodes, err := g.nodeInfoSnapshot.NodeInfos().List()
	if err != nil {
		klog.Errorf("node inf snapshot list nodes error %v", err)
		return nil, err
	}

	numNodesToFind := g.numFeasibleNodesToFind(int32(len(allNodes)))
	klog.Infof("Number of feasible node to find is %d, all nodes number is %d", numNodesToFind, len(allNodes))

	feasibleNodes := make([]*v1.Node, numNodesToFind)

	if !prof.HasFilterPlugins() {
		length := len(allNodes)
		for i := range feasibleNodes {
			feasibleNodes[i] = allNodes[(g.nextStartNodeIndex+i)%length].Node()
		}
		g.nextStartNodeIndex = (g.nextStartNodeIndex + len(feasibleNodes)) % length
		return feasibleNodes, nil
	}

	errCh := parallelize.NewErrorChannel()
	var statusesLock sync.Mutex
	var feasibleNodesLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		nodeInfo := allNodes[(g.nextStartNodeIndex+i)%len(allNodes)]
		fits, status, err := PodPassesFiltersOnNode(ctx, prof.PreemptHandle(), state, pod, nodeInfo)
		if err != nil {
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
		if fits {
			length := atomic.AddInt32(&feasibleNodesLen, 1)
			if length > numNodesToFind {
				cancel()
				atomic.AddInt32(&feasibleNodesLen, -1)
			} else {
				feasibleNodes[length-1] = nodeInfo.Node()
			}
		} else {
			statusesLock.Lock()
			if !status.IsSuccess() {
				statuses[nodeInfo.Node().Name] = status
			}
			statusesLock.Unlock()
		}
	}
	parallelize.Until(ctx, len(allNodes), checkNode)
	processedNodes := int(feasibleNodesLen) + len(statuses)
	g.nextStartNodeIndex = (g.nextStartNodeIndex + processedNodes) % len(allNodes)

	feasibleNodes = feasibleNodes[:feasibleNodesLen]
	if err := errCh.ReceiveError(); err != nil {
		return nil, err
	}
	return feasibleNodes, nil
}

func addNominatedPods(ctx context.Context, ph framework.PreemptHandle, pod *v1.Pod, state *framework.CycleState, nodeInfo *framework.NodeInfo) (bool, *framework.CycleState, *framework.NodeInfo, error) {
	if ph == nil || nodeInfo == nil || nodeInfo.Node() == nil {
		return false, state, nodeInfo, nil
	}
	nominatedPods := ph.NominatedPodsForNode(nodeInfo.Node().Name)
	if len(nominatedPods) == 0 {
		return false, state, nodeInfo, nil
	}
	nodeInfoOut := nodeInfo.Clone()
	stateOut := state.Clone()
	podsAdded := false
	for _, p := range nominatedPods {
		if podutil.GetPodPriority(p) >= podutil.GetPodPriority(pod) && p.UID != pod.UID {
			nodeInfoOut.AddPod(p)
			status := ph.RunPreFilterExtensionAddPod(ctx, stateOut, pod, p, nodeInfoOut)
			if !status.IsSuccess() {
				return false, state, nodeInfo, status.AsError()
			}
			podsAdded = true
		}
	}
	return podsAdded, stateOut, nodeInfoOut, nil
}

func PodPassesFiltersOnNode(
	ctx context.Context,
	ph framework.PreemptHandle,
	state *framework.CycleState,
	pod *v1.Pod,
	info *framework.NodeInfo,
) (bool, *framework.Status, error) {
	var status *framework.Status

	podsAdded := false
	for i := 0; i < 2; i++ {
		stateToUse := state
		nodeInfoToUse := info
		if i == 0 {
			var err error
			podsAdded, stateToUse, nodeInfoToUse, err = addNominatedPods(ctx, ph, pod, state, info)
			if err != nil {
				return false, nil, err
			}
		} else if !podsAdded || !status.IsSuccess() {
			break
		}

		statusMap := ph.RunFilterPlugins(ctx, stateToUse, pod, nodeInfoToUse)
		status = statusMap.Merge()
		if !status.IsSuccess() && !status.IsUnschedulable() {
			return false, status, status.AsError()
		}
	}

	return status.IsSuccess(), status, nil
}

func (g *genericScheduler) prioritizeNodes(
	ctx context.Context,
	prof *profile.Profile,
	state *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (framework.NodeScoreList, error) {
	if !prof.HasScorePlugins() {
		result := make(framework.NodeScoreList, 0, len(nodes))
		for i := range nodes {
			result = append(result, framework.NodeScore{
				Name:  nodes[i].Name,
				Score: 1,
			})
		}
		return result, nil
	}

	preScoreStatus := prof.RunPreScorePlugins(ctx, state, pod, nodes)
	if !preScoreStatus.IsSuccess() {
		return nil, preScoreStatus.AsError()
	}

	scoresMap, scoreStatus := prof.RunScorePlugins(ctx, state, pod, nodes)
	if !scoreStatus.IsSuccess() {
		return nil, scoreStatus.AsError()
	}

	if klog.V(10).Enabled() {
		for plugin, nodeScoreList := range scoresMap {
			klog.Infof("Plugin %s scores on %v/%v => %v", plugin, pod.Namespace, pod.Name, nodeScoreList)
		}
	}

	result := make(framework.NodeScoreList, 0, len(nodes))

	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i].Name, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	if false {
		for i := range result {
			klog.Infof("Host %s => Score %d", result[i].Name, result[i].Score)
		}
	}
	return result, nil
}

func podPassesBasicChecks(pod *v1.Pod) error {
	return nil
}

func NewGenericScheduler(
	cache internalcache.Cache,
	nodeInfoSnapshot *internalcache.Snapshot,
	percentageOfNodesToScore int32) ScheduleAlgorithm {
	return &genericScheduler{
		cache:                    cache,
		nodeInfoSnapshot:         nodeInfoSnapshot,
		percentageOfNodesToScore: percentageOfNodesToScore,
	}
}
