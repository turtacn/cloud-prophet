package helper

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// DefaultSelector returns a selector deduced from the Services, Replication
// Controllers, Replica Sets, and Stateful Sets matching the given pod.
func DefaultSelector(pod *v1.Pod, sl corelisters.ServiceLister, cl corelisters.ReplicationControllerLister, rsl appslisters.ReplicaSetLister, ssl appslisters.StatefulSetLister) labels.Selector {
	labelSet := make(labels.Set)
	// Since services, RCs, RSs and SSs match the pod, they won't have conflicting
	// labels. Merging is safe.

	if services, err := GetPodServices(sl, pod); err == nil {
		for _, service := range services {
			labelSet = labels.Merge(labelSet, service.Spec.Selector)
		}
	}

	if rcs, err := cl.GetPodControllers(pod); err == nil {
		for _, rc := range rcs {
			labelSet = labels.Merge(labelSet, rc.Spec.Selector)
		}
	}

	selector := labels.NewSelector()
	if len(labelSet) != 0 {
		selector = labelSet.AsSelector()
	}

	if rss, err := rsl.GetPodReplicaSets(pod); err == nil {
		for _, rs := range rss {
			if other, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector); err == nil {
				if r, ok := other.Requirements(); ok {
					selector = selector.Add(r...)
				}
			}
		}
	}

	if sss, err := ssl.GetPodStatefulSets(pod); err == nil {
		for _, ss := range sss {
			if other, err := metav1.LabelSelectorAsSelector(ss.Spec.Selector); err == nil {
				if r, ok := other.Requirements(); ok {
					selector = selector.Add(r...)
				}
			}
		}
	}

	return selector
}

// GetPodServices gets the services that have the selector that match the labels on the given pod.
func GetPodServices(sl corelisters.ServiceLister, pod *v1.Pod) ([]*v1.Service, error) {
	allServices, err := sl.Services(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var services []*v1.Service
	for i := range allServices {
		service := allServices[i]
		if service.Spec.Selector == nil {
			// services with nil selectors match nothing, not everything.
			continue
		}
		selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(pod.Labels)) {
			services = append(services, service)
		}
	}

	return services, nil
}
