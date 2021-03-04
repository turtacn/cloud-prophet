//
//
package runtime

import (
	"fmt"
	"sync"
	"time"

	k8s "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

// waitingPodsMap a thread-safe map used to maintain pods waiting in the permit phase.
type waitingPodsMap struct {
	pods map[string]*waitingPod
	mu   sync.RWMutex
}

// newWaitingPodsMap returns a new waitingPodsMap.
func newWaitingPodsMap() *waitingPodsMap {
	return &waitingPodsMap{
		pods: make(map[string]*waitingPod),
	}
}

// add a new WaitingPod to the map.
func (m *waitingPodsMap) add(wp *waitingPod) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pods[wp.GetPod().UID] = wp
}

// remove a WaitingPod from the map.
func (m *waitingPodsMap) remove(uid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pods, uid)
}

// get a WaitingPod from the map.
func (m *waitingPodsMap) get(uid string) *waitingPod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pods[uid]
}

// iterate acquires a read lock and iterates over the WaitingPods map.
func (m *waitingPodsMap) iterate(callback func(k8s.WaitingPod)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.pods {
		callback(v)
	}
}

// waitingPod represents a pod waiting in the permit phase.
type waitingPod struct {
	pod            *v1.Pod
	pendingPlugins map[string]*time.Timer
	s              chan *k8s.Status
	mu             sync.RWMutex
}

var _ k8s.WaitingPod = &waitingPod{}

// newWaitingPod returns a new waitingPod instance.
func newWaitingPod(pod *v1.Pod, pluginsMaxWaitTime map[string]time.Duration) *waitingPod {
	wp := &waitingPod{
		pod: pod,
		// Allow() and Reject() calls are non-blocking. This property is guaranteed
		// by using non-blocking send to this channel. This channel has a buffer of size 1
		// to ensure that non-blocking send will not be ignored - possible situation when
		// receiving from this channel happens after non-blocking send.
		s: make(chan *k8s.Status, 1),
	}

	wp.pendingPlugins = make(map[string]*time.Timer, len(pluginsMaxWaitTime))
	// The time.AfterFunc calls wp.Reject which iterates through pendingPlugins map. Acquire the
	// lock here so that time.AfterFunc can only execute after newWaitingPod finishes.
	wp.mu.Lock()
	defer wp.mu.Unlock()
	for k, v := range pluginsMaxWaitTime {
		plugin, waitTime := k, v
		wp.pendingPlugins[plugin] = time.AfterFunc(waitTime, func() {
			msg := fmt.Sprintf("rejected due to timeout after waiting %v at plugin %v",
				waitTime, plugin)
			wp.Reject(msg)
		})
	}

	return wp
}

// GetPod returns a reference to the waiting pod.
func (w *waitingPod) GetPod() *v1.Pod {
	return w.pod
}

// GetPendingPlugins returns a list of pending permit plugin's name.
func (w *waitingPod) GetPendingPlugins() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	plugins := make([]string, 0, len(w.pendingPlugins))
	for p := range w.pendingPlugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// Allow declares the waiting pod is allowed to be scheduled by plugin pluginName.
// If this is the last remaining plugin to allow, then a success signal is delivered
// to unblock the pod.
func (w *waitingPod) Allow(pluginName string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if timer, exist := w.pendingPlugins[pluginName]; exist {
		timer.Stop()
		delete(w.pendingPlugins, pluginName)
	}

	// Only signal success status after all plugins have allowed
	if len(w.pendingPlugins) != 0 {
		return
	}

	// The select clause works as a non-blocking send.
	// If there is no receiver, it's a no-op (default case).
	select {
	case w.s <- k8s.NewStatus(k8s.Success, ""):
	default:
	}
}

// Reject declares the waiting pod unschedulable.
func (w *waitingPod) Reject(msg string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, timer := range w.pendingPlugins {
		timer.Stop()
	}

	// The select clause works as a non-blocking send.
	// If there is no receiver, it's a no-op (default case).
	select {
	case w.s <- k8s.NewStatus(k8s.Unschedulable, msg):
	default:
	}
}
