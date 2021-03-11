package queue

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/klog/v2"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	metav1 "github.com/turtacn/cloud-prophet/scheduler/helper"
	"github.com/turtacn/cloud-prophet/scheduler/internal/heap"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/util"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	unschedulableQTimeInterval = 60 * time.Second

	queueClosed = "scheduling queue is closed"
)

const (
	DefaultPodInitialBackoffDuration time.Duration = 1 * time.Second
	DefaultPodMaxBackoffDuration     time.Duration = 10 * time.Second
)

type SchedulingQueue interface {
	framework.PodNominator
	Add(pod *v1.Pod) error
	AddUnschedulableIfNotPresent(pod *framework.QueuedPodInfo, podSchedulingCycle int64) error
	SchedulingCycle() int64
	Pop() (*framework.QueuedPodInfo, error)
	Update(oldPod, newPod *v1.Pod) error
	Delete(pod *v1.Pod) error
	MoveAllToActiveOrBackoffQueue(event string)
	AssignedPodAdded(pod *v1.Pod)
	AssignedPodUpdated(pod *v1.Pod)
	PendingPods() []*v1.Pod
	Close()
	NumUnschedulablePods() int
	Run()
}

func NewSchedulingQueue(lessFn framework.LessFunc, opts ...Option) SchedulingQueue {
	return NewPriorityQueue(lessFn, opts...)
}

func NominatedNodeName(pod *v1.Pod) string {
	return pod.Status.NominatedNodeName
}

type PriorityQueue struct {
	framework.PodNominator

	stop  chan struct{}
	clock util.Clock

	podInitialBackoffDuration time.Duration
	podMaxBackoffDuration     time.Duration

	lock sync.RWMutex
	cond sync.Cond

	activeQ          *heap.Heap
	podBackoffQ      *heap.Heap
	unschedulableQ   *UnschedulablePodsMap
	schedulingCycle  int64
	moveRequestCycle int64

	closed bool
}

type priorityQueueOptions struct {
	clock                     util.Clock
	podInitialBackoffDuration time.Duration
	podMaxBackoffDuration     time.Duration
	podNominator              framework.PodNominator
}

type Option func(*priorityQueueOptions)

func WithClock(clock util.Clock) Option {
	return func(o *priorityQueueOptions) {
		o.clock = clock
	}
}

func WithPodInitialBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.podInitialBackoffDuration = duration
	}
}

func WithPodMaxBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.podMaxBackoffDuration = duration
	}
}

func WithPodNominator(pn framework.PodNominator) Option {
	return func(o *priorityQueueOptions) {
		o.podNominator = pn
	}
}

var defaultPriorityQueueOptions = priorityQueueOptions{
	clock:                     util.RealClock{},
	podInitialBackoffDuration: DefaultPodInitialBackoffDuration,
	podMaxBackoffDuration:     DefaultPodMaxBackoffDuration,
}

var _ SchedulingQueue = &PriorityQueue{}

func newQueuedPodInfoNoTimestamp(pod *v1.Pod) *framework.QueuedPodInfo {
	return &framework.QueuedPodInfo{
		Pod: pod,
	}
}

func NewPriorityQueue(
	lessFn framework.LessFunc,
	opts ...Option,
) *PriorityQueue {
	options := defaultPriorityQueueOptions
	for _, opt := range opts {
		opt(&options)
	}

	comp := func(podInfo1, podInfo2 interface{}) bool {
		pInfo1 := podInfo1.(*framework.QueuedPodInfo)
		pInfo2 := podInfo2.(*framework.QueuedPodInfo)
		return lessFn(pInfo1, pInfo2)
	}

	if options.podNominator == nil {
		options.podNominator = NewPodNominator()
	}

	pq := &PriorityQueue{
		PodNominator:              options.podNominator,
		clock:                     options.clock,
		stop:                      make(chan struct{}),
		podInitialBackoffDuration: options.podInitialBackoffDuration,
		podMaxBackoffDuration:     options.podMaxBackoffDuration,
		activeQ:                   heap.New(podInfoKeyFunc, comp),
		unschedulableQ:            newUnschedulablePodsMap(),
		moveRequestCycle:          -1,
	}
	pq.cond.L = &pq.lock
	pq.podBackoffQ = heap.New(podInfoKeyFunc, pq.podsCompareBackoffCompleted)

	return pq
}

func (p *PriorityQueue) Run() {
	go wait.Until(p.flushBackoffQCompleted, 1.0*time.Second, p.stop)
	go wait.Until(p.flushUnschedulableQLeftover, 30*time.Second, p.stop)
}

func (p *PriorityQueue) Add(pod *v1.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	pInfo := p.newQueuedPodInfo(pod)
	if err := p.activeQ.Add(pInfo); err != nil {
		klog.Errorf("Error adding pod %v to the scheduling queue: %v", nsNameForPod(pod), err)
		return err
	}
	if p.unschedulableQ.get(pod) != nil {
		klog.Errorf("Error: pod %v is already in the unschedulable queue.", nsNameForPod(pod))
		p.unschedulableQ.delete(pod)
	}
	if err := p.podBackoffQ.Delete(pInfo); err == nil {
		klog.Errorf("Error: pod %v is already in the podBackoff queue.", nsNameForPod(pod))
	}
	p.PodNominator.AddNominatedPod(pod, "")
	p.cond.Broadcast()

	return nil
}

func nsNameForPod(pod *v1.Pod) ktypes.NamespacedName {
	return ktypes.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}
}

func (p *PriorityQueue) isPodBackingoff(podInfo *framework.QueuedPodInfo) bool {
	boTime := p.getBackoffTime(podInfo)
	return boTime.After(p.clock.Now())
}

func (p *PriorityQueue) SchedulingCycle() int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.schedulingCycle
}

func (p *PriorityQueue) AddUnschedulableIfNotPresent(pInfo *framework.QueuedPodInfo, podSchedulingCycle int64) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	pod := pInfo.Pod
	if p.unschedulableQ.get(pod) != nil {
		return fmt.Errorf("pod: %v is already present in unschedulable queue", nsNameForPod(pod))
	}

	pInfo.Timestamp = p.clock.Now()
	if _, exists, _ := p.activeQ.Get(pInfo); exists {
		return fmt.Errorf("pod: %v is already present in the active queue", nsNameForPod(pod))
	}
	if _, exists, _ := p.podBackoffQ.Get(pInfo); exists {
		return fmt.Errorf("pod %v is already present in the backoff queue", nsNameForPod(pod))
	}

	if p.moveRequestCycle >= podSchedulingCycle {
		if err := p.podBackoffQ.Add(pInfo); err != nil {
			return fmt.Errorf("error adding pod %v to the backoff queue: %v", pod.Name, err)
		}
	} else {
		p.unschedulableQ.addOrUpdate(pInfo)
	}

	p.PodNominator.AddNominatedPod(pod, "")
	return nil
}

func (p *PriorityQueue) flushBackoffQCompleted() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for {
		rawPodInfo := p.podBackoffQ.Peek()
		if rawPodInfo == nil {
			return
		}
		pod := rawPodInfo.(*framework.QueuedPodInfo).Pod
		boTime := p.getBackoffTime(rawPodInfo.(*framework.QueuedPodInfo))
		if boTime.After(p.clock.Now()) {
			return
		}
		_, err := p.podBackoffQ.Pop()
		if err != nil {
			klog.Errorf("Unable to pop pod %v from backoff queue despite backoff completion.", nsNameForPod(pod))
			return
		}
		p.activeQ.Add(rawPodInfo)
		defer p.cond.Broadcast()
	}
}

func (p *PriorityQueue) flushUnschedulableQLeftover() {
	p.lock.Lock()
	defer p.lock.Unlock()

	var podsToMove []*framework.QueuedPodInfo
	currentTime := p.clock.Now()
	for _, pInfo := range p.unschedulableQ.podInfoMap {
		lastScheduleTime := pInfo.Timestamp
		if currentTime.Sub(lastScheduleTime) > unschedulableQTimeInterval {
			podsToMove = append(podsToMove, pInfo)
		}
	}

	if len(podsToMove) > 0 {
		p.movePodsToActiveOrBackoffQueue(podsToMove, UnschedulableTimeout)
	}
}

func (p *PriorityQueue) Pop() (*framework.QueuedPodInfo, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for p.activeQ.Len() == 0 {
		if p.closed {
			return nil, fmt.Errorf(queueClosed)
		}
		p.cond.Wait()
	}
	obj, err := p.activeQ.Pop()
	if err != nil {
		return nil, err
	}
	pInfo := obj.(*framework.QueuedPodInfo)
	pInfo.Attempts++
	p.schedulingCycle++
	return pInfo, err
}

func isPodUpdated(oldPod, newPod *v1.Pod) bool {
	strip := func(pod *v1.Pod) *v1.Pod {
		p := pod.DeepCopy()
		p.ResourceVersion = ""
		p.Generation = 0
		p.Status = v1.PodStatus{}
		return p
	}
	return !reflect.DeepEqual(strip(oldPod), strip(newPod))
}

func (p *PriorityQueue) Update(oldPod, newPod *v1.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if oldPod != nil {
		oldPodInfo := newQueuedPodInfoNoTimestamp(oldPod)
		if oldPodInfo, exists, _ := p.activeQ.Get(oldPodInfo); exists {
			p.PodNominator.UpdateNominatedPod(oldPod, newPod)
			err := p.activeQ.Update(updatePod(oldPodInfo, newPod))
			return err
		}

		if oldPodInfo, exists, _ := p.podBackoffQ.Get(oldPodInfo); exists {
			p.PodNominator.UpdateNominatedPod(oldPod, newPod)
			p.podBackoffQ.Delete(oldPodInfo)
			err := p.activeQ.Add(updatePod(oldPodInfo, newPod))
			if err == nil {
				p.cond.Broadcast()
			}
			return err
		}
	}

	if usPodInfo := p.unschedulableQ.get(newPod); usPodInfo != nil {
		p.PodNominator.UpdateNominatedPod(oldPod, newPod)
		if isPodUpdated(oldPod, newPod) {
			p.unschedulableQ.delete(usPodInfo.Pod)
			err := p.activeQ.Add(updatePod(usPodInfo, newPod))
			if err == nil {
				p.cond.Broadcast()
			}
			return err
		}
		p.unschedulableQ.addOrUpdate(updatePod(usPodInfo, newPod))
		return nil
	}
	err := p.activeQ.Add(p.newQueuedPodInfo(newPod))
	if err == nil {
		p.PodNominator.AddNominatedPod(newPod, "")
		p.cond.Broadcast()
	}
	return err
}

func (p *PriorityQueue) Delete(pod *v1.Pod) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.PodNominator.DeleteNominatedPodIfExists(pod)
	err := p.activeQ.Delete(newQueuedPodInfoNoTimestamp(pod))
	if err != nil { // The item was probably not found in the activeQ.
		p.podBackoffQ.Delete(newQueuedPodInfoNoTimestamp(pod))
		p.unschedulableQ.delete(pod)
	}
	return nil
}

func (p *PriorityQueue) AssignedPodAdded(pod *v1.Pod) {
	p.lock.Lock()
	p.movePodsToActiveOrBackoffQueue(p.getUnschedulablePodsWithMatchingAffinityTerm(pod), AssignedPodAdd)
	p.lock.Unlock()
}

func (p *PriorityQueue) AssignedPodUpdated(pod *v1.Pod) {
	p.lock.Lock()
	p.movePodsToActiveOrBackoffQueue(p.getUnschedulablePodsWithMatchingAffinityTerm(pod), AssignedPodUpdate)
	p.lock.Unlock()
}

func (p *PriorityQueue) MoveAllToActiveOrBackoffQueue(event string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	unschedulablePods := make([]*framework.QueuedPodInfo, 0, len(p.unschedulableQ.podInfoMap))
	for _, pInfo := range p.unschedulableQ.podInfoMap {
		unschedulablePods = append(unschedulablePods, pInfo)
	}
	p.movePodsToActiveOrBackoffQueue(unschedulablePods, event)
}

func (p *PriorityQueue) movePodsToActiveOrBackoffQueue(podInfoList []*framework.QueuedPodInfo, event string) {
	for _, pInfo := range podInfoList {
		pod := pInfo.Pod
		if p.isPodBackingoff(pInfo) {
			if err := p.podBackoffQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding pod %v to the backoff queue: %v", pod.Name, err)
			} else {
				p.unschedulableQ.delete(pod)
			}
		} else {
			if err := p.activeQ.Add(pInfo); err != nil {
				klog.Errorf("Error adding pod %v to the scheduling queue: %v", pod.Name, err)
			} else {
				p.unschedulableQ.delete(pod)
			}
		}
	}
	p.moveRequestCycle = p.schedulingCycle
	p.cond.Broadcast()
}

func (p *PriorityQueue) getUnschedulablePodsWithMatchingAffinityTerm(pod *v1.Pod) []*framework.QueuedPodInfo {
	var podsToMove []*framework.QueuedPodInfo
	for _, pInfo := range p.unschedulableQ.podInfoMap {
		up := pInfo.Pod
		terms := util.GetPodAffinityTerms(up.Spec.Affinity)
		for _, term := range terms {
			namespaces := util.GetNamespacesFromPodAffinityTerm(up, &term)
			selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
			if err != nil {
				klog.Errorf("Error getting label selectors for pod: %v.", up.Name)
			}
			if util.PodMatchesTermsNamespaceAndSelector(pod, namespaces, selector) {
				podsToMove = append(podsToMove, pInfo)
				break
			}
		}

	}
	return podsToMove
}

func (p *PriorityQueue) PendingPods() []*v1.Pod {
	p.lock.RLock()
	defer p.lock.RUnlock()
	var result []*v1.Pod
	for _, pInfo := range p.activeQ.List() {
		result = append(result, pInfo.(*framework.QueuedPodInfo).Pod)
	}
	for _, pInfo := range p.podBackoffQ.List() {
		result = append(result, pInfo.(*framework.QueuedPodInfo).Pod)
	}
	for _, pInfo := range p.unschedulableQ.podInfoMap {
		result = append(result, pInfo.Pod)
	}
	return result
}

func (p *PriorityQueue) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()
	close(p.stop)
	p.closed = true
	p.cond.Broadcast()
}

func (npm *nominatedPodMap) DeleteNominatedPodIfExists(pod *v1.Pod) {
	npm.Lock()
	npm.delete(pod)
	npm.Unlock()
}

func (npm *nominatedPodMap) AddNominatedPod(pod *v1.Pod, nodeName string) {
	npm.Lock()
	npm.add(pod, nodeName)
	npm.Unlock()
}

func (npm *nominatedPodMap) NominatedPodsForNode(nodeName string) []*v1.Pod {
	npm.RLock()
	defer npm.RUnlock()
	return npm.nominatedPods[nodeName]
}

func (p *PriorityQueue) podsCompareBackoffCompleted(podInfo1, podInfo2 interface{}) bool {
	pInfo1 := podInfo1.(*framework.QueuedPodInfo)
	pInfo2 := podInfo2.(*framework.QueuedPodInfo)
	bo1 := p.getBackoffTime(pInfo1)
	bo2 := p.getBackoffTime(pInfo2)
	return bo1.Before(bo2)
}

func (p *PriorityQueue) NumUnschedulablePods() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.unschedulableQ.podInfoMap)
}

func (p *PriorityQueue) newQueuedPodInfo(pod *v1.Pod) *framework.QueuedPodInfo {
	now := p.clock.Now()
	return &framework.QueuedPodInfo{
		Pod:                     pod,
		Timestamp:               now,
		InitialAttemptTimestamp: now,
	}
}

func (p *PriorityQueue) getBackoffTime(podInfo *framework.QueuedPodInfo) time.Time {
	duration := p.calculateBackoffDuration(podInfo)
	backoffTime := podInfo.Timestamp.Add(duration)
	return backoffTime
}

func (p *PriorityQueue) calculateBackoffDuration(podInfo *framework.QueuedPodInfo) time.Duration {
	duration := p.podInitialBackoffDuration
	for i := 1; i < podInfo.Attempts; i++ {
		duration = duration * 2
		if duration > p.podMaxBackoffDuration {
			return p.podMaxBackoffDuration
		}
	}
	return duration
}

func updatePod(oldPodInfo interface{}, newPod *v1.Pod) *framework.QueuedPodInfo {
	pInfo := oldPodInfo.(*framework.QueuedPodInfo)
	pInfo.Pod = newPod
	return pInfo
}

type UnschedulablePodsMap struct {
	podInfoMap map[string]*framework.QueuedPodInfo
	keyFunc    func(*v1.Pod) string
}

func (u *UnschedulablePodsMap) addOrUpdate(pInfo *framework.QueuedPodInfo) {
	podID := u.keyFunc(pInfo.Pod)
	u.podInfoMap[podID] = pInfo
}

func (u *UnschedulablePodsMap) delete(pod *v1.Pod) {
	podID := u.keyFunc(pod)
	delete(u.podInfoMap, podID)
}

func (u *UnschedulablePodsMap) get(pod *v1.Pod) *framework.QueuedPodInfo {
	podKey := u.keyFunc(pod)
	if pInfo, exists := u.podInfoMap[podKey]; exists {
		return pInfo
	}
	return nil
}

func (u *UnschedulablePodsMap) clear() {
	u.podInfoMap = make(map[string]*framework.QueuedPodInfo)
}

func newUnschedulablePodsMap() *UnschedulablePodsMap {
	return &UnschedulablePodsMap{
		podInfoMap: make(map[string]*framework.QueuedPodInfo),
		keyFunc:    util.GetPodFullName,
	}
}

type nominatedPodMap struct {
	nominatedPods      map[string][]*v1.Pod
	nominatedPodToNode map[string]string

	sync.RWMutex
}

func (npm *nominatedPodMap) add(p *v1.Pod, nodeName string) {
	npm.delete(p)

	nnn := nodeName
	if len(nnn) == 0 {
		nnn = NominatedNodeName(p)
		if len(nnn) == 0 {
			return
		}
	}
	npm.nominatedPodToNode[p.UID] = nnn
	for _, np := range npm.nominatedPods[nnn] {
		if np.UID == p.UID {
			klog.Infof("Pod %v/%v already exists in the nominated map!", p.Namespace, p.Name)
			return
		}
	}
	npm.nominatedPods[nnn] = append(npm.nominatedPods[nnn], p)
}

func (npm *nominatedPodMap) delete(p *v1.Pod) {
	nnn, ok := npm.nominatedPodToNode[p.UID]
	if !ok {
		return
	}
	for i, np := range npm.nominatedPods[nnn] {
		if np.UID == p.UID {
			npm.nominatedPods[nnn] = append(npm.nominatedPods[nnn][:i], npm.nominatedPods[nnn][i+1:]...)
			if len(npm.nominatedPods[nnn]) == 0 {
				delete(npm.nominatedPods, nnn)
			}
			break
		}
	}
	delete(npm.nominatedPodToNode, p.UID)
}

func (npm *nominatedPodMap) UpdateNominatedPod(oldPod, newPod *v1.Pod) {
	npm.Lock()
	defer npm.Unlock()
	nodeName := ""
	if NominatedNodeName(oldPod) == "" && NominatedNodeName(newPod) == "" {
		if nnn, ok := npm.nominatedPodToNode[oldPod.UID]; ok {
			nodeName = nnn
		}
	}
	npm.delete(oldPod)
	npm.add(newPod, nodeName)
}

func NewPodNominator() framework.PodNominator {
	return &nominatedPodMap{
		nominatedPods:      make(map[string][]*v1.Pod),
		nominatedPodToNode: make(map[string]string),
	}
}

func MakeNextPodFunc(queue SchedulingQueue) func() *framework.QueuedPodInfo {
	return func() *framework.QueuedPodInfo {
		podInfo, err := queue.Pop()
		if err == nil {
			klog.Infof("About to try and schedule pod %v/%v", podInfo.Pod.Namespace, podInfo.Pod.Name)
			return podInfo
		}
		klog.Errorf("Error while retrieving next pod from scheduling queue: %v", err)
		return nil
	}
}

func podInfoKeyFunc(obj interface{}) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.(*framework.QueuedPodInfo).Pod.UID, nil
}
