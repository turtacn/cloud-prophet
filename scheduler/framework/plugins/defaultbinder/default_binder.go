package defaultbinder

import (
	"context"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Bind binds pods to nodes using the k8s client.
func (b DefaultBinder) Bind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) *framework.Status {
	klog.V(3).Infof("Attempting to bind %v/%v to %v", p.Namespace, p.Name, nodeName)
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: p.Namespace, Name: p.Name, UID: p.UID},
		Target:     v1.ObjectReference{Kind: "Node", Name: nodeName},
	}
	err := b.handle.ClientSet().CoreV1().Pods(binding.Namespace).Bind(ctx, binding, metav1.CreateOptions{})
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	return nil
}
