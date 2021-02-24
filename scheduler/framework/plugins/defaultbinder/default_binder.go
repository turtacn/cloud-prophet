//
//
package defaultbinder

import (
	"context"

	"errors"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/k8s"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Name of the plugin used in the plugin registry and configurations.
const Name = "DefaultBinder"

// DefaultBinder binds pods to nodes using a k8s client.
type DefaultBinder struct {
	handle framework.FrameworkHandle
}

var _ framework.BindPlugin = &DefaultBinder{}

// New creates a DefaultBinder.
func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &DefaultBinder{handle: handle}, nil
}

// Name returns the name of the plugin.
func (b DefaultBinder) Name() string {
	return Name
}

// Bind binds pods to nodes using the k8s/jvirt client.
func (b DefaultBinder) Bind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	klog.Infof("Attempting to bind %v/%v to %v", p.Namespace, p.Name, nodeName)
	binding := &v1.Binding{
		ObjectMeta: v1.ObjectMeta{Namespace: p.Namespace, Name: p.Name, UID: p.UID},
		Target:     v1.ObjectReference{Kind: "Node", Name: nodeName},
	}
	node, e := b.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if e != nil {
		return framework.NewStatus(framework.Error, e.Error())
	}
	node.AddPod(p)
	for i, _ := range node.Pods {
		klog.Infof("pod %s binding node %s has pods: ", p.Name, nodeName, node.Pods[i])
	}

	return nil

	var err error = errors.New("xxx")
	if b.handle.ClientSet() == nil || binding == nil {

	}
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	return nil
}
