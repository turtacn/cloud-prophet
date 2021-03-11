//
package runtime

import (
	"fmt"
	"sync"
	"time"

	k8s "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

type waitingPodsMap struct {
	pods map[string]*waitingPod
	mu   sync.RWMutex
}

func newWaitingPodsMap() *waitingPodsMap {
	return &waitingPodsMap{
		pods: make(map[string]*waitingPod),
	}
}

func (m *waitingPodsMap) add(wp *waitingPod) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pods[wp.GetPod().UID] = wp
}

func (m *waitingPodsMap) remove(uid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pods, uid)
}

func (m *waitingPodsMap) get(uid string) *waitingPod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pods[uid]
}

func (m *waitingPodsMap) iterate(callback func(k8s.WaitingPod)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.pods {
		callback(v)
	}
}

type waitingPod struct {
	pod            *v1.Pod
	pendingPlugins map[string]*time.Timer
	s              chan *k8s.Status
	mu             sync.RWMutex
}

var _ k8s.WaitingPod = &waitingPod{}

func newWaitingPod(pod *v1.Pod, pluginsMaxWaitTime map[string]time.Duration) *waitingPod {
	wp := &waitingPod{
		pod: pod,
		s:   make(chan *k8s.Status, 1),
	}

	wp.pendingPlugins = make(map[string]*time.Timer, len(pluginsMaxWaitTime))
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

func (w *waitingPod) GetPod() *v1.Pod {
	return w.pod
}

func (w *waitingPod) GetPendingPlugins() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	plugins := make([]string, 0, len(w.pendingPlugins))
	for p := range w.pendingPlugins {
		plugins = append(plugins, p)
	}

	return plugins
}

func (w *waitingPod) Allow(pluginName string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if timer, exist := w.pendingPlugins[pluginName]; exist {
		timer.Stop()
		delete(w.pendingPlugins, pluginName)
	}

	if len(w.pendingPlugins) != 0 {
		return
	}

	select {
	case w.s <- k8s.NewStatus(k8s.Success, ""):
	default:
	}
}

func (w *waitingPod) Reject(msg string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, timer := range w.pendingPlugins {
		timer.Stop()
	}

	select {
	case w.s <- k8s.NewStatus(k8s.Unschedulable, msg):
	default:
	}
}
