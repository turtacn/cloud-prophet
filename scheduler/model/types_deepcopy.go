//
package model

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (in *Node) DeepCopy() *Node {
	return nil
}
func (in *TopologySpreadConstraint) DeepCopyInto(out *TopologySpreadConstraint) {

}

func (in *Pod) DeepCopyInto(out *Pod) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	return
}

func (in *Pod) DeepCopy() (out *Pod) {
	if in == nil {
		return nil
	}
	out = new(Pod)
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
	return nil
}
