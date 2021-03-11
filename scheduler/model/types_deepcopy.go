package model

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (in *Node) DeepCopyInto(out *Node) {
	*out = *in
	out.TypeMeta = in.TypeMeta
}
func (in *Node) DeepCopy() *Node {
	if in == nil {
		return nil
	}
	out := new(Node)
	in.DeepCopyInto(out)
	return out
}
func (in *Node) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *TopologySpreadConstraint) DeepCopyInto(out *TopologySpreadConstraint) {

}

func (in *Pod) DeepCopyInto(out *Pod) {
	*out = *in
	out.TypeMeta = in.TypeMeta
}

func (in *Pod) DeepCopy() *Pod {
	if in == nil {
		return nil
	}
	out := new(Pod)
	in.DeepCopyInto(out)
	return out
}

func (in *Pod) DeepCopyObject() runtime.Object {

	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *Pod) GetObjectKind() schema.ObjectKind {
	return in
}
