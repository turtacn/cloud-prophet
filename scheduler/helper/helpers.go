package helper

import (
	"fmt"
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	"github.com/turtacn/cloud-prophet/scheduler/model"
)

// LabelSelectorAsSelector converts the LabelSelector api type into a struct that implements
// labels.Selector
// Note: This function should be kept in sync with the selector methods in pkg/labels/selector.go
func LabelSelectorAsSelector(ps *model.LabelSelector) (labels.Selector, error) {
	if ps == nil {
		return labels.Nothing(), nil
	}
	if len(ps.MatchLabels)+len(ps.MatchExpressions) == 0 {
		return labels.Everything(), nil
	}
	selector := labels.NewSelector()
	for k, v := range ps.MatchLabels {
		r, err := labels.NewRequirement(k, labels.Equals, []string{v})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	for _, expr := range ps.MatchExpressions {
		var op labels.Operator
		switch expr.Operator {
		case model.LabelSelectorOpIn:
			op = labels.In
		case model.LabelSelectorOpNotIn:
			op = labels.NotIn
		case model.LabelSelectorOpExists:
			op = labels.Exists
		case model.LabelSelectorOpDoesNotExist:
			op = labels.DoesNotExist
		default:
			return nil, fmt.Errorf("%q is not a valid pod selector operator", expr.Operator)
		}
		r, err := labels.NewRequirement(expr.Key, op, append([]string(nil), expr.Values...))
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}
