package config

import (
	"github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kube-scheduler/config/v1beta1"
)

const GroupName = v1beta1.GroupName

var SchemeGroupVersion = v1beta1.SchemeGroupVersion

var (
	localSchemeBuilder = &v1beta1.SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(RegisterDefaults)
}

func RegisterDefaults(scheme *runtime.Scheme) error {

	scheme.AddTypeDefaultingFunc(&PodTopologySpreadArgs{}, func(obj interface{}) {
		SetObjectDefaults_PodTopologySpreadArgs(obj.(*PodTopologySpreadArgs))
	})
	return nil
}

func SetObjectDefaults_PodTopologySpreadArgs(in *PodTopologySpreadArgs) {
	var obj = in
	if obj.DefaultConstraints == nil {
		obj.DefaultConstraints = []model.TopologySpreadConstraint{
			{
				MaxSkew:           3,
				TopologyKey:       "",
				WhenUnsatisfiable: "",
			},
		}
	}
}
