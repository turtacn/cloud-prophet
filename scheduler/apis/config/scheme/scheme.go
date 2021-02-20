//
//
package scheme

import (
	kubeschedulerconfig "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	// Scheme is the runtime.Scheme to which all kubescheduler api types are registered.
	Scheme = runtime.NewScheme()

	// Codecs provides access to encoding and decoding for the scheme.
	Codecs = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)
)

func init() {
	AddToScheme(Scheme)
}

func AddToScheme(scheme *runtime.Scheme) {
}

func NewFromSchemeByName(name string) runtime.Object {
	switch {
	case name == "PodTopologySpread":
		return &kubeschedulerconfig.PodTopologySpreadArgs{}
	case name == "NodeResourcesLeastAllocated":
		return &kubeschedulerconfig.NodeResourcesLeastAllocatedArgs{}
	case name == "InterPodAffinity":
		return &kubeschedulerconfig.InterPodAffinityArgs{}
	case name == "NodeResourcesFit":
		return &kubeschedulerconfig.NodeResourcesFitArgs{}
	default:
		return nil
	}
	return nil
}
