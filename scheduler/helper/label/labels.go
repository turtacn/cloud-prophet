package label

import (
	"fmt"
	"sort"
	"strings"
)

type Labels interface {
	Has(label string) (exists bool)

	Get(label string) (value string)
}

type Set map[string]string

func (ls Set) String() string {
	selector := make([]string, 0, len(ls))
	for key, value := range ls {
		selector = append(selector, key+"="+value)
	}
	sort.StringSlice(selector).Sort()
	return strings.Join(selector, ",")
}

func (ls Set) Has(label string) bool {
	_, exists := ls[label]
	return exists
}

func (ls Set) Get(label string) string {
	return ls[label]
}

func (ls Set) AsSelector() Selector {
	return SelectorFromSet(ls)
}

func (ls Set) AsValidatedSelector() (Selector, error) {
	return ValidatedSelectorFromSet(ls)
}

func (ls Set) AsSelectorPreValidated() Selector {
	return SelectorFromValidatedSet(ls)
}

func FormatLabels(labelMap map[string]string) string {
	l := Set(labelMap).String()
	if l == "" {
		l = "<none>"
	}
	return l
}

func Conflicts(labels1, labels2 Set) bool {
	small := labels1
	big := labels2
	if len(labels2) < len(labels1) {
		small = labels2
		big = labels1
	}

	for k, v := range small {
		if val, match := big[k]; match {
			if val != v {
				return true
			}
		}
	}

	return false
}

func Merge(labels1, labels2 Set) Set {
	mergedMap := Set{}

	for k, v := range labels1 {
		mergedMap[k] = v
	}
	for k, v := range labels2 {
		mergedMap[k] = v
	}
	return mergedMap
}

func Equal(labels1, labels2 Set) bool {
	if len(labels1) != len(labels2) {
		return false
	}

	for k, v := range labels1 {
		value, ok := labels2[k]
		if !ok {
			return false
		}
		if value != v {
			return false
		}
	}
	return true
}

func AreLabelsInWhiteList(labels, whitelist Set) bool {
	if len(whitelist) == 0 {
		return true
	}

	for k, v := range labels {
		value, ok := whitelist[k]
		if !ok {
			return false
		}
		if value != v {
			return false
		}
	}
	return true
}

func ConvertSelectorToLabelsMap(selector string) (Set, error) {
	labelsMap := Set{}

	if len(selector) == 0 {
		return labelsMap, nil
	}

	labels := strings.Split(selector, ",")
	for _, label := range labels {
		l := strings.Split(label, "=")
		if len(l) != 2 {
			return labelsMap, fmt.Errorf("invalid selector: %s", l)
		}
		key := strings.TrimSpace(l[0])
		if err := validateLabelKey(key); err != nil {
			return labelsMap, err
		}
		value := strings.TrimSpace(l[1])
		if err := validateLabelValue(key, value); err != nil {
			return labelsMap, err
		}
		labelsMap[key] = value
	}
	return labelsMap, nil
}
