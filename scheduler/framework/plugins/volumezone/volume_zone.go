package volumezone

import (
	"context"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	corelisters "k8s.io/client-go/listers/core/v1"
	storagelisters "k8s.io/client-go/listers/storage/v1"
)

// VolumeZone is a plugin that checks volume zone.
type VolumeZone struct {
	pvLister  corelisters.PersistentVolumeLister
	pvcLister corelisters.PersistentVolumeClaimLister
	scLister  storagelisters.StorageClassLister
}

var _ framework.FilterPlugin = &VolumeZone{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "VolumeZone"

	// ErrReasonConflict is used for NoVolumeZoneConflict predicate error.
	ErrReasonConflict = "node(s) had no available volume zone"
)

var volumeZoneLabels = sets.NewString(
	v1.LabelZoneFailureDomain,
	v1.LabelZoneRegion,
	v1.LabelZoneFailureDomainStable,
	v1.LabelZoneRegionStable,
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *VolumeZone) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
//
// It evaluates if a pod can fit due to the volumes it requests, given
// that some volumes may have zone scheduling constraints.  The requirement is that any
// volume zone-labels must match the equivalent zone-labels on the node.  It is OK for
// the node to have more zone-label constraints (for example, a hypothetical replicated
// volume might allow region-wide access)
//
// Currently this is only supported with PersistentVolumeClaims, and looks to the labels
// only on the bound PersistentVolume.
//
// Working with volumes declared inline in the pod specification (i.e. not
// using a PersistentVolume) is likely to be harder, as it would require
// determining the zone of a volume during scheduling, and that is likely to
// require calling out to the cloud provider.  It seems that we are moving away
// from inline volume declarations anyway.
func (pl *VolumeZone) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	// If a pod doesn't have any volume attached to it, the predicate will always be true.
	// Thus we make a fast path for it, to avoid unnecessary computations in this case.
	if len(pod.Spec.Volumes) == 0 {
		return nil
	}
	node := nodeInfo.Node()
	if node == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}
	nodeConstraints := make(map[string]string)
	for k, v := range node.ObjectMeta.Labels {
		if !volumeZoneLabels.Has(k) {
			continue
		}
		nodeConstraints[k] = v
	}
	if len(nodeConstraints) == 0 {
		// The node has no zone constraints, so we're OK to schedule.
		// In practice, when using zones, all nodes must be labeled with zone labels.
		// We want to fast-path this case though.
		return nil
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	informerFactory := handle.SharedInformerFactory()
	pvLister := informerFactory.Core().V1().PersistentVolumes().Lister()
	pvcLister := informerFactory.Core().V1().PersistentVolumeClaims().Lister()
	scLister := informerFactory.Storage().V1().StorageClasses().Lister()
	return &VolumeZone{
		pvLister,
		pvcLister,
		scLister,
	}, nil
}
