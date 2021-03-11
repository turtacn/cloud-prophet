package base

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	metav1 "github.com/turtacn/cloud-prophet/scheduler/helper"
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	schedutil "github.com/turtacn/cloud-prophet/scheduler/util"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

var generation int64

type QueuedPodInfo struct {
	Pod                     *v1.Pod
	Timestamp               time.Time
	Attempts                int
	InitialAttemptTimestamp time.Time
}

func (pqi *QueuedPodInfo) DeepCopy() *QueuedPodInfo {
	return &QueuedPodInfo{
		Pod:                     pqi.Pod.DeepCopy(),
		Timestamp:               pqi.Timestamp,
		Attempts:                pqi.Attempts,
		InitialAttemptTimestamp: pqi.InitialAttemptTimestamp,
	}
}

type PodInfo struct {
	Pod                        *v1.Pod
	RequiredAffinityTerms      []AffinityTerm
	RequiredAntiAffinityTerms  []AffinityTerm
	PreferredAffinityTerms     []WeightedAffinityTerm
	PreferredAntiAffinityTerms []WeightedAffinityTerm
	ParseError                 error
}

type AffinityTerm struct {
	Namespaces  sets.String
	Selector    labels.Selector
	TopologyKey string
}

type WeightedAffinityTerm struct {
	AffinityTerm
	Weight int32
}

func newAffinityTerm(pod *v1.Pod, term *v1.PodAffinityTerm) (*AffinityTerm, error) {
	namespaces := schedutil.GetNamespacesFromPodAffinityTerm(pod, term)
	selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
	if err != nil {
		return nil, err
	}
	return &AffinityTerm{Namespaces: namespaces, Selector: selector, TopologyKey: term.TopologyKey}, nil
}

func getAffinityTerms(pod *v1.Pod, v1Terms []v1.PodAffinityTerm) ([]AffinityTerm, error) {
	if v1Terms == nil {
		return nil, nil
	}

	var terms []AffinityTerm
	for _, term := range v1Terms {
		t, err := newAffinityTerm(pod, &term)
		if err != nil {
			return nil, err
		}
		terms = append(terms, *t)
	}
	return terms, nil
}

func getWeightedAffinityTerms(pod *v1.Pod, v1Terms []v1.WeightedPodAffinityTerm) ([]WeightedAffinityTerm, error) {
	if v1Terms == nil {
		return nil, nil
	}

	var terms []WeightedAffinityTerm
	for _, term := range v1Terms {
		t, err := newAffinityTerm(pod, &term.PodAffinityTerm)
		if err != nil {
			return nil, err
		}
		terms = append(terms, WeightedAffinityTerm{AffinityTerm: *t, Weight: term.Weight})
	}
	return terms, nil
}

func NewPodInfo(pod *v1.Pod) *PodInfo {
	var preferredAffinityTerms []v1.WeightedPodAffinityTerm
	var preferredAntiAffinityTerms []v1.WeightedPodAffinityTerm
	if affinity := pod.Spec.Affinity; affinity != nil {
		if a := affinity.PodAffinity; a != nil {
			preferredAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
		}
		if a := affinity.PodAntiAffinity; a != nil {
			preferredAntiAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
		}
	}

	var parseErr error
	requiredAffinityTerms, err := getAffinityTerms(pod, schedutil.GetPodAffinityTerms(pod.Spec.Affinity))
	if err != nil {
		parseErr = fmt.Errorf("requiredAffinityTerms: %w", err)
	}
	requiredAntiAffinityTerms, err := getAffinityTerms(pod, schedutil.GetPodAntiAffinityTerms(pod.Spec.Affinity))
	if err != nil {
		parseErr = fmt.Errorf("requiredAntiAffinityTerms: %w", err)
	}
	weightedAffinityTerms, err := getWeightedAffinityTerms(pod, preferredAffinityTerms)
	if err != nil {
		parseErr = fmt.Errorf("preferredAffinityTerms: %w", err)
	}
	weightedAntiAffinityTerms, err := getWeightedAffinityTerms(pod, preferredAntiAffinityTerms)
	if err != nil {
		parseErr = fmt.Errorf("preferredAntiAffinityTerms: %w", err)
	}

	return &PodInfo{
		Pod:                        pod,
		RequiredAffinityTerms:      requiredAffinityTerms,
		RequiredAntiAffinityTerms:  requiredAntiAffinityTerms,
		PreferredAffinityTerms:     weightedAffinityTerms,
		PreferredAntiAffinityTerms: weightedAntiAffinityTerms,
		ParseError:                 parseErr,
	}
}

type ImageStateSummary struct {
	Size     int64
	NumNodes int
}

type NodeInfo struct {
	node *v1.Node

	Pods []*PodInfo

	PodsWithAffinity []*PodInfo

	Requested        *Resource
	NonZeroRequested *Resource
	Allocatable      *Resource

	ImageStates map[string]*ImageStateSummary

	Generation int64
}

func nextGeneration() int64 {
	return atomic.AddInt64(&generation, 1)
}

type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
	AllowedPodNumber int
	ScalarResources  map[v1.ResourceName]int64
}

func NewResource(rl v1.ResourceList) *Resource {
	r := &Resource{}
	r.Add(rl)
	return r
}

func (r *Resource) Add(rl v1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuant := range rl {
		switch rName {
		case v1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case v1.ResourceMemory:
			r.Memory += rQuant.Value()
		case v1.ResourcePods:
			r.AllowedPodNumber += int(rQuant.Value())
		case v1.ResourceEphemeralStorage:
			r.EphemeralStorage += rQuant.Value()
		default:
			r.AddScalar(rName, rQuant.Value())
		}
	}
}

func (r *Resource) ResourceList() v1.ResourceList {
	result := v1.ResourceList{
		v1.ResourceCPU:              *resource.NewMilliQuantity(r.MilliCPU, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(r.Memory, resource.BinarySI),
		v1.ResourcePods:             *resource.NewQuantity(int64(r.AllowedPodNumber), resource.BinarySI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(r.EphemeralStorage, resource.BinarySI),
	}
	for rName, rQuant := range r.ScalarResources {
		result[rName] = *resource.NewQuantity(rQuant, resource.BinarySI)
	}
	return result
}

func (r *Resource) Clone() *Resource {
	res := &Resource{
		MilliCPU:         r.MilliCPU,
		Memory:           r.Memory,
		AllowedPodNumber: r.AllowedPodNumber,
		EphemeralStorage: r.EphemeralStorage,
	}
	if r.ScalarResources != nil {
		res.ScalarResources = make(map[v1.ResourceName]int64)
		for k, v := range r.ScalarResources {
			res.ScalarResources[k] = v
		}
	}
	return res
}

func (r *Resource) AddScalar(name v1.ResourceName, quantity int64) {
	r.SetScalar(name, r.ScalarResources[name]+quantity)
}

func (r *Resource) SetScalar(name v1.ResourceName, quantity int64) {
	if r.ScalarResources == nil {
		r.ScalarResources = map[v1.ResourceName]int64{}
	}
	r.ScalarResources[name] = quantity
}

func (r *Resource) SetMaxResource(rl v1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuantity := range rl {
		switch rName {
		case v1.ResourceMemory:
			if mem := rQuantity.Value(); mem > r.Memory {
				r.Memory = mem
			}
		case v1.ResourceCPU:
			if cpu := rQuantity.MilliValue(); cpu > r.MilliCPU {
				r.MilliCPU = cpu
			}
		case v1.ResourceEphemeralStorage:
			if ephemeralStorage := rQuantity.Value(); ephemeralStorage > r.EphemeralStorage {
				r.EphemeralStorage = ephemeralStorage
			}
		default:
			value := rQuantity.Value()
			if value > r.ScalarResources[rName] {
				r.SetScalar(rName, value)
			}
		}
	}
}

func NewNodeInfo(pods ...*v1.Pod) *NodeInfo {
	ni := &NodeInfo{
		Requested:        &Resource{},
		NonZeroRequested: &Resource{},
		Allocatable:      &Resource{},
		Generation:       nextGeneration(),
		ImageStates:      make(map[string]*ImageStateSummary),
	}
	for _, pod := range pods {
		ni.AddPod(pod)
	}
	return ni
}

func (n *NodeInfo) Node() *v1.Node {
	if n == nil {
		return nil
	}
	return n.node
}

func (n *NodeInfo) Clone() *NodeInfo {
	clone := &NodeInfo{
		node:             n.node,
		Requested:        n.Requested.Clone(),
		NonZeroRequested: n.NonZeroRequested.Clone(),
		Allocatable:      n.Allocatable.Clone(),
		ImageStates:      n.ImageStates,
		Generation:       n.Generation,
	}
	if len(n.Pods) > 0 {
		clone.Pods = append([]*PodInfo(nil), n.Pods...)
	}
	if len(n.PodsWithAffinity) > 0 {
		clone.PodsWithAffinity = append([]*PodInfo(nil), n.PodsWithAffinity...)
	}
	return clone
}

func (n *NodeInfo) String() string {
	podKeys := make([]string, len(n.Pods))
	for i, p := range n.Pods {
		podKeys[i] = p.Pod.Name
	}
	return fmt.Sprintf("&NodeInfo{Pods:%v, RequestedResource:%#v, NonZeroRequest: %#v, AllocatableResource:%#v}",
		podKeys, n.Requested, n.NonZeroRequested, n.Allocatable)
}

func (n *NodeInfo) AddPod(pod *v1.Pod) {
	podInfo := NewPodInfo(pod)
	res, non0CPU, non0Mem := calculateResource(pod)
	n.Requested.MilliCPU += res.MilliCPU
	n.Requested.Memory += res.Memory
	n.Requested.EphemeralStorage += res.EphemeralStorage
	if n.Requested.ScalarResources == nil && len(res.ScalarResources) > 0 {
		n.Requested.ScalarResources = map[v1.ResourceName]int64{}
	}
	for rName, rQuant := range res.ScalarResources {
		n.Requested.ScalarResources[rName] += rQuant
	}
	n.NonZeroRequested.MilliCPU += non0CPU
	n.NonZeroRequested.Memory += non0Mem
	n.Pods = append(n.Pods, podInfo)
	affinity := pod.Spec.Affinity
	if affinity != nil && (affinity.PodAffinity != nil || affinity.PodAntiAffinity != nil) {
		n.PodsWithAffinity = append(n.PodsWithAffinity, podInfo)
	}

	n.Generation = nextGeneration()
}

func (n *NodeInfo) RemovePod(pod *v1.Pod) error {
	k1, err := GetPodKey(pod)
	if err != nil {
		return err
	}

	for i := range n.PodsWithAffinity {
		k2, err := GetPodKey(n.PodsWithAffinity[i].Pod)
		if err != nil {
			klog.Errorf("Cannot get pod key, err: %v", err)
			continue
		}
		if k1 == k2 {
			n.PodsWithAffinity[i] = n.PodsWithAffinity[len(n.PodsWithAffinity)-1]
			n.PodsWithAffinity = n.PodsWithAffinity[:len(n.PodsWithAffinity)-1]
			break
		}
	}
	for i := range n.Pods {
		k2, err := GetPodKey(n.Pods[i].Pod)
		if err != nil {
			klog.Errorf("Cannot get pod key, err: %v", err)
			continue
		}
		if k1 == k2 {
			n.Pods[i] = n.Pods[len(n.Pods)-1]
			n.Pods = n.Pods[:len(n.Pods)-1]
			res, non0CPU, non0Mem := calculateResource(pod)

			n.Requested.MilliCPU -= res.MilliCPU
			n.Requested.Memory -= res.Memory
			n.Requested.EphemeralStorage -= res.EphemeralStorage
			if len(res.ScalarResources) > 0 && n.Requested.ScalarResources == nil {
				n.Requested.ScalarResources = map[v1.ResourceName]int64{}
			}
			for rName, rQuant := range res.ScalarResources {
				n.Requested.ScalarResources[rName] -= rQuant
			}
			n.NonZeroRequested.MilliCPU -= non0CPU
			n.NonZeroRequested.Memory -= non0Mem

			n.Generation = nextGeneration()
			n.resetSlicesIfEmpty()
			return nil
		}
	}
	return fmt.Errorf("no corresponding pod %s in pods of node %s", pod.Name, n.node.Name)
}

func (n *NodeInfo) resetSlicesIfEmpty() {
	if len(n.PodsWithAffinity) == 0 {
		n.PodsWithAffinity = nil
	}
	if len(n.Pods) == 0 {
		n.Pods = nil
	}
}

func calculateResource(pod *v1.Pod) (res Resource, non0CPU int64, non0Mem int64) {
	resPtr := &res
	for _, c := range pod.Spec.Containers {
		resPtr.Add(c.Resources.Requests)
		non0CPUReq, non0MemReq := schedutil.GetNonzeroRequests(&c.Resources.Requests)
		non0CPU += non0CPUReq
		non0Mem += non0MemReq
	}

	if pod.Spec.Overhead != nil {
		resPtr.Add(pod.Spec.Overhead)
		if _, found := pod.Spec.Overhead[v1.ResourceCPU]; found {
			non0CPU += pod.Spec.Overhead.Cpu().MilliValue()
		}

		if _, found := pod.Spec.Overhead[v1.ResourceMemory]; found {
			non0Mem += pod.Spec.Overhead.Memory().Value()
		}
	}

	return
}

func (n *NodeInfo) updateUsedPorts(pod *v1.Pod, add bool) {
}

func (n *NodeInfo) SetNode(node *v1.Node) error {
	n.node = node
	n.Allocatable = NewResource(node.Status.Allocatable)
	n.Generation = nextGeneration()
	return nil
}

func (n *NodeInfo) RemoveNode() {
	n.node = nil
	n.Generation = nextGeneration()
}

func (n *NodeInfo) FilterOutPods(pods []*v1.Pod) []*v1.Pod {
	node := n.Node()
	if node == nil {
		return pods
	}
	filtered := make([]*v1.Pod, 0, len(pods))
	for _, p := range pods {
		if p.Spec.NodeName != node.Name {
			filtered = append(filtered, p)
			continue
		}
		podKey, err := GetPodKey(p)
		if err != nil {
			continue
		}
		for _, np := range n.Pods {
			npodkey, _ := GetPodKey(np.Pod)
			if npodkey == podKey {
				filtered = append(filtered, p)
				break
			}
		}
	}
	return filtered
}

func GetPodKey(pod *v1.Pod) (string, error) {
	uid := string(pod.UID)
	if len(uid) == 0 {
		return "", errors.New("Cannot get cache key for pod with empty UID")
	}
	return uid, nil
}

const DefaultBindAllHostIP = "0.0.0.0"

type ProtocolPort struct {
	Protocol string
	Port     int32
}
