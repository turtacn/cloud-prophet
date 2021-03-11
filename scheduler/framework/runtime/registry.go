//
package runtime

import (
	"fmt"

	k8s "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"
)

type PluginFactory = func(configuration runtime.Object, f k8s.FrameworkHandle) (k8s.Plugin, error)

func DecodeInto(obj runtime.Object, into interface{}) error {
	if obj == nil {
		return nil
	}
	configuration, ok := obj.(*runtime.Unknown)
	if !ok {
		return fmt.Errorf("want args of type runtime.Unknown, got %T", obj)
	}
	if configuration.Raw == nil {
		return nil
	}

	switch configuration.ContentType {
	case runtime.ContentTypeJSON, "":
		return json.Unmarshal(configuration.Raw, into)
	case runtime.ContentTypeYAML:
		return yaml.Unmarshal(configuration.Raw, into)
	default:
		return fmt.Errorf("not supported content type %s", configuration.ContentType)
	}
}

type Registry map[string]PluginFactory

func (r Registry) Register(name string, factory PluginFactory) error {
	if _, ok := r[name]; ok {
		return fmt.Errorf("a plugin named %v already exists", name)
	}
	r[name] = factory
	return nil
}

func (r Registry) Unregister(name string) error {
	if _, ok := r[name]; !ok {
		return fmt.Errorf("no plugin named %v exists", name)
	}
	delete(r, name)
	return nil
}

func (r Registry) Merge(in Registry) error {
	for name, factory := range in {
		if err := r.Register(name, factory); err != nil {
			return err
		}
	}
	return nil
}
