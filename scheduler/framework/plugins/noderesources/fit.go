package noderesources

import (
	"context"
	"fmt"
	"strings"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ framework.PreFilterPlugin = &Fit{}
var _ framework.FilterPlugin = &Fit{}

const (
	FitName = "NodeResourcesFit"

	preFilterStateKey = "PreFilter" + FitName
)

type Fit struct {
	ignoredResources      sets.String
	ignoredResourceGroups sets.String
}

type preFilterState struct {
	framework.Resource
}

func (s *preFilterState) Clone() framework.StateData {
	return s
}

func (f *Fit) Name() string {
	return FitName
}

func validateFitArgs(args config.NodeResourcesFitArgs) error {
	var allErrs field.ErrorList
	resPath := field.NewPath("ignoredResources")
	for i, _ := range args.IgnoredResources {
		resPath.Index(i)
	}

	groupPath := field.NewPath("ignoredResourceGroups")
	for i, group := range args.IgnoredResourceGroups {
		path := groupPath.Index(i)
		if strings.Contains(group, "/") {
			allErrs = append(allErrs, field.Invalid(path, group, "resource group name can't contain '/'"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}

func NewFit(plArgs runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getFitArgs(plArgs)
	if err != nil {
		return nil, err
	}

	if err := validateFitArgs(args); err != nil {
		return nil, err
	}

	return &Fit{
		ignoredResources:      sets.NewString(args.IgnoredResources...),
		ignoredResourceGroups: sets.NewString(args.IgnoredResourceGroups...),
	}, nil
}

func getFitArgs(obj runtime.Object) (config.NodeResourcesFitArgs, error) {
	ptr, ok := obj.(*config.NodeResourcesFitArgs)
	if !ok {
		return config.NodeResourcesFitArgs{}, fmt.Errorf("want args to be of type NodeResourcesFitArgs, got %T", obj)
	}
	return *ptr, nil
}

func computePodResourceRequest(pod *v1.Pod) *preFilterState {
	result := &preFilterState{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	if pod.Spec.Overhead != nil {
		result.Add(pod.Spec.Overhead)
	}

	return result
}

func (f *Fit) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) *framework.Status {
	cycleState.Write(preFilterStateKey, computePodResourceRequest(pod))
	return nil
}

func (f *Fit) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func getPreFilterState(cycleState *framework.CycleState) (*preFilterState, error) {
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to NodeResourcesFit.preFilterState error", c)
	}
	return s, nil
}

func (f *Fit) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	s, err := getPreFilterState(cycleState)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	insufficientResources := fitsRequest(s, nodeInfo, f.ignoredResources, f.ignoredResourceGroups)

	if len(insufficientResources) != 0 {
		failureReasons := make([]string, 0, len(insufficientResources))
		for _, r := range insufficientResources {
			failureReasons = append(failureReasons, r.Reason)
		}
		return framework.NewStatus(framework.Unschedulable, failureReasons...)
	}
	return nil
}

type InsufficientResource struct {
	ResourceName v1.ResourceName
	Reason       string
	Requested    int64
	Used         int64
	Capacity     int64
}

func Fits(pod *v1.Pod, nodeInfo *framework.NodeInfo) []InsufficientResource {
	return fitsRequest(computePodResourceRequest(pod), nodeInfo, nil, nil)
}

func fitsRequest(podRequest *preFilterState, nodeInfo *framework.NodeInfo, ignoredExtendedResources, ignoredResourceGroups sets.String) []InsufficientResource {
	insufficientResources := make([]InsufficientResource, 0, 4)

	allowedPodNumber := nodeInfo.Allocatable.AllowedPodNumber
	if len(nodeInfo.Pods)+1 > allowedPodNumber {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourcePods,
			"Too many pods",
			1,
			int64(len(nodeInfo.Pods)),
			int64(allowedPodNumber),
		})
	}

	if podRequest.MilliCPU == 0 &&
		podRequest.Memory == 0 &&
		podRequest.EphemeralStorage == 0 &&
		len(podRequest.ScalarResources) == 0 {
		return insufficientResources
	}

	if nodeInfo.Allocatable.MilliCPU < podRequest.MilliCPU+nodeInfo.Requested.MilliCPU {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceCPU,
			"Insufficient cpu",
			podRequest.MilliCPU,
			nodeInfo.Requested.MilliCPU,
			nodeInfo.Allocatable.MilliCPU,
		})
	}
	if nodeInfo.Allocatable.Memory < podRequest.Memory+nodeInfo.Requested.Memory {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceMemory,
			"Insufficient memory",
			podRequest.Memory,
			nodeInfo.Requested.Memory,
			nodeInfo.Allocatable.Memory,
		})
	}
	if nodeInfo.Allocatable.EphemeralStorage < podRequest.EphemeralStorage+nodeInfo.Requested.EphemeralStorage {
		insufficientResources = append(insufficientResources, InsufficientResource{
			v1.ResourceEphemeralStorage,
			"Insufficient ephemeral-storage",
			podRequest.EphemeralStorage,
			nodeInfo.Requested.EphemeralStorage,
			nodeInfo.Allocatable.EphemeralStorage,
		})
	}

	for rName, rQuant := range podRequest.ScalarResources {
		var rNamePrefix string
		if ignoredResourceGroups.Len() > 0 {
			rNamePrefix = strings.Split(string(rName), "/")[0]
		}
		if ignoredExtendedResources.Has(string(rName)) || ignoredResourceGroups.Has(rNamePrefix) {
			continue
		}
		if nodeInfo.Allocatable.ScalarResources[rName] < rQuant+nodeInfo.Requested.ScalarResources[rName] {
			insufficientResources = append(insufficientResources, InsufficientResource{
				rName,
				fmt.Sprintf("Insufficient %v", rName),
				podRequest.ScalarResources[rName],
				nodeInfo.Requested.ScalarResources[rName],
				nodeInfo.Allocatable.ScalarResources[rName],
			})
		}
	}

	return insufficientResources
}
