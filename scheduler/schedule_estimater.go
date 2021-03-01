package scheduler

import (
	"github.com/turtacn/cloud-prophet/scheduler/model"
)

func (s *Scheduler) AddNode(node *model.Node) {
	s.addNodeToCache(node)
}
func (s *Scheduler) DeleteNode(node *model.Node) {
	s.deleteNodeFromCache(node)
}
func (s *Scheduler) UpdateNode(old, new *model.Node) {
	s.updateNodeInCache(old, new)
}
func (s *Scheduler) AddPod(pod *model.Pod) {
	s.addPodToCache(pod)
}
func (s *Scheduler) DeletePod(pod *model.Pod) {
	s.deletePodFromCache(pod)
}
func (s *Scheduler) UpdatePod(old, new *model.Pod) {
	s.updatePodInCache(old, new)
}
