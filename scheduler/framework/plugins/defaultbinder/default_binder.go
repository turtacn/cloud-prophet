//
package defaultbinder

import (
	"context"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const Name = "DefaultBinder"

type DefaultBinder struct {
	handle framework.FrameworkHandle
}

var _ framework.BindPlugin = &DefaultBinder{}

func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &DefaultBinder{handle: handle}, nil
}

func (b DefaultBinder) Name() string {
	return Name
}

func (b DefaultBinder) Bind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	klog.Infof("Attempting to bind %v/%v to %v", p.Namespace, p.Name, nodeName)
	binding := &v1.Binding{
		ObjectMeta: v1.ObjectMeta{Namespace: p.Namespace, Name: p.Name, UID: p.UID},
		Target:     v1.ObjectReference{Kind: "Node", Name: nodeName},
	}
	if binding == nil {
	}
	return nil
}
