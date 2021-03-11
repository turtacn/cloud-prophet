//
package config

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

func (in *Extender) DeepCopyInto(out *Extender) {
	*out = *in
	if in.TLSConfig != nil {
		in, out := &in.TLSConfig, &out.TLSConfig
		*out = new(ExtenderTLSConfig)
		(*in).DeepCopyInto(*out)
	}
	out.HTTPTimeout = in.HTTPTimeout
	if in.ManagedResources != nil {
		in, out := &in.ManagedResources, &out.ManagedResources
		*out = make([]ExtenderManagedResource, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *Extender) DeepCopy() *Extender {
	if in == nil {
		return nil
	}
	out := new(Extender)
	in.DeepCopyInto(out)
	return out
}

func (in *ExtenderManagedResource) DeepCopyInto(out *ExtenderManagedResource) {
	*out = *in
	return
}

func (in *ExtenderManagedResource) DeepCopy() *ExtenderManagedResource {
	if in == nil {
		return nil
	}
	out := new(ExtenderManagedResource)
	in.DeepCopyInto(out)
	return out
}

func (in *ExtenderTLSConfig) DeepCopyInto(out *ExtenderTLSConfig) {
	*out = *in
	if in.CertData != nil {
		in, out := &in.CertData, &out.CertData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.KeyData != nil {
		in, out := &in.KeyData, &out.KeyData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.CAData != nil {
		in, out := &in.CAData, &out.CAData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *ExtenderTLSConfig) DeepCopy() *ExtenderTLSConfig {
	if in == nil {
		return nil
	}
	out := new(ExtenderTLSConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *InterPodAffinityArgs) DeepCopyInto(out *InterPodAffinityArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	return
}

func (in *InterPodAffinityArgs) DeepCopy() *InterPodAffinityArgs {
	if in == nil {
		return nil
	}
	out := new(InterPodAffinityArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *InterPodAffinityArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *KubeSchedulerProfile) DeepCopyInto(out *KubeSchedulerProfile) {
	*out = *in
	if in.Plugins != nil {
		in, out := &in.Plugins, &out.Plugins
		*out = new(Plugins)
		(*in).DeepCopyInto(*out)
	}
	if in.PluginConfig != nil {
		in, out := &in.PluginConfig, &out.PluginConfig
		*out = make([]PluginConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

func (in *KubeSchedulerProfile) DeepCopy() *KubeSchedulerProfile {
	if in == nil {
		return nil
	}
	out := new(KubeSchedulerProfile)
	in.DeepCopyInto(out)
	return out
}

func (in *NodeResourcesFitArgs) DeepCopyInto(out *NodeResourcesFitArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.IgnoredResources != nil {
		in, out := &in.IgnoredResources, &out.IgnoredResources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IgnoredResourceGroups != nil {
		in, out := &in.IgnoredResourceGroups, &out.IgnoredResourceGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *NodeResourcesFitArgs) DeepCopy() *NodeResourcesFitArgs {
	if in == nil {
		return nil
	}
	out := new(NodeResourcesFitArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *NodeResourcesFitArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *NodeResourcesLeastAllocatedArgs) DeepCopyInto(out *NodeResourcesLeastAllocatedArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceSpec, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *NodeResourcesLeastAllocatedArgs) DeepCopy() *NodeResourcesLeastAllocatedArgs {
	if in == nil {
		return nil
	}
	out := new(NodeResourcesLeastAllocatedArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *NodeResourcesLeastAllocatedArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *NodeResourcesMostAllocatedArgs) DeepCopyInto(out *NodeResourcesMostAllocatedArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceSpec, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *NodeResourcesMostAllocatedArgs) DeepCopy() *NodeResourcesMostAllocatedArgs {
	if in == nil {
		return nil
	}
	out := new(NodeResourcesMostAllocatedArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *NodeResourcesMostAllocatedArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *Plugin) DeepCopyInto(out *Plugin) {
	*out = *in
	return
}

func (in *Plugin) DeepCopy() *Plugin {
	if in == nil {
		return nil
	}
	out := new(Plugin)
	in.DeepCopyInto(out)
	return out
}

func (in *PluginConfig) DeepCopyInto(out *PluginConfig) {
	*out = *in
	if in.Args != nil {
		out.Args = in.Args.DeepCopyObject()
	}
	return
}

func (in *PluginConfig) DeepCopy() *PluginConfig {
	if in == nil {
		return nil
	}
	out := new(PluginConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *PluginSet) DeepCopyInto(out *PluginSet) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = make([]Plugin, len(*in))
		copy(*out, *in)
	}
	if in.Disabled != nil {
		in, out := &in.Disabled, &out.Disabled
		*out = make([]Plugin, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *PluginSet) DeepCopy() *PluginSet {
	if in == nil {
		return nil
	}
	out := new(PluginSet)
	in.DeepCopyInto(out)
	return out
}

func (in *Plugins) DeepCopyInto(out *Plugins) {
	*out = *in
	if in.QueueSort != nil {
		in, out := &in.QueueSort, &out.QueueSort
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.PreFilter != nil {
		in, out := &in.PreFilter, &out.PreFilter
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Filter != nil {
		in, out := &in.Filter, &out.Filter
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.PostFilter != nil {
		in, out := &in.PostFilter, &out.PostFilter
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.PreScore != nil {
		in, out := &in.PreScore, &out.PreScore
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Score != nil {
		in, out := &in.Score, &out.Score
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Reserve != nil {
		in, out := &in.Reserve, &out.Reserve
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Permit != nil {
		in, out := &in.Permit, &out.Permit
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.PreBind != nil {
		in, out := &in.PreBind, &out.PreBind
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.Bind != nil {
		in, out := &in.Bind, &out.Bind
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	if in.PostBind != nil {
		in, out := &in.PostBind, &out.PostBind
		*out = new(PluginSet)
		(*in).DeepCopyInto(*out)
	}
	return
}

func (in *Plugins) DeepCopy() *Plugins {
	if in == nil {
		return nil
	}
	out := new(Plugins)
	in.DeepCopyInto(out)
	return out
}

func (in *PodTopologySpreadArgs) DeepCopyInto(out *PodTopologySpreadArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.DefaultConstraints != nil {
		in, out := &in.DefaultConstraints, &out.DefaultConstraints
		*out = make([]v1.TopologySpreadConstraint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

func (in *PodTopologySpreadArgs) DeepCopy() *PodTopologySpreadArgs {
	if in == nil {
		return nil
	}
	out := new(PodTopologySpreadArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *PodTopologySpreadArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *RequestedToCapacityRatioArgs) DeepCopyInto(out *RequestedToCapacityRatioArgs) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Shape != nil {
		in, out := &in.Shape, &out.Shape
		*out = make([]UtilizationShapePoint, len(*in))
		copy(*out, *in)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceSpec, len(*in))
		copy(*out, *in)
	}
	return
}

func (in *RequestedToCapacityRatioArgs) DeepCopy() *RequestedToCapacityRatioArgs {
	if in == nil {
		return nil
	}
	out := new(RequestedToCapacityRatioArgs)
	in.DeepCopyInto(out)
	return out
}

func (in *RequestedToCapacityRatioArgs) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ResourceSpec) DeepCopyInto(out *ResourceSpec) {
	*out = *in
	return
}

func (in *ResourceSpec) DeepCopy() *ResourceSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *SchedulerAlgorithmSource) DeepCopyInto(out *SchedulerAlgorithmSource) {
	*out = *in
	if in.Provider != nil {
		in, out := &in.Provider, &out.Provider
		*out = new(string)
		**out = **in
	}
	return
}

func (in *SchedulerAlgorithmSource) DeepCopy() *SchedulerAlgorithmSource {
	if in == nil {
		return nil
	}
	out := new(SchedulerAlgorithmSource)
	in.DeepCopyInto(out)
	return out
}

func (in *UtilizationShapePoint) DeepCopyInto(out *UtilizationShapePoint) {
	*out = *in
	return
}

func (in *UtilizationShapePoint) DeepCopy() *UtilizationShapePoint {
	if in == nil {
		return nil
	}
	out := new(UtilizationShapePoint)
	in.DeepCopyInto(out)
	return out
}
