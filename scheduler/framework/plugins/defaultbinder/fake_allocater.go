//
//
package defaultbinder

import (
	"context"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/k8s"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Name of the plugin used in the plugin registry and configurations.
const NameFakeAllocater = "FakeAllocater"

// DefaultBinder binds pods to nodes using a k8s client.
type FakeAllocater struct {
	handle framework.FrameworkHandle
}

var _ framework.PostBindPlugin = &FakeAllocater{}

// New creates a DefaultBinder.
func NewFA(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &FakeAllocater{handle: handle}, nil
}

// Name returns the name of the plugin.
func (b FakeAllocater) Name() string {
	return NameFakeAllocater
}

func (b FakeAllocater) PostBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) {
	klog.Infof("Attempting to post bind %v/%v to %v", p.Namespace, p.Name, nodeName)
}
