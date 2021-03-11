package model

type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms" protobuf:"bytes,1,rep,name=nodeSelectorTerms"`
}

type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,1,rep,name=matchExpressions"`
	MatchFields      []NodeSelectorRequirement `json:"matchFields,omitempty" protobuf:"bytes,2,rep,name=matchFields"`
}

type NodeSelectorRequirement struct {
	Key      string               `json:"key" protobuf:"bytes,1,opt,name=key"`
	Operator NodeSelectorOperator `json:"operator" protobuf:"bytes,2,opt,name=operator,casttype=NodeSelectorOperator"`
	Values   []string             `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"`
}

type NodeSelectorOperator string

const (
	NodeSelectorOpIn           NodeSelectorOperator = "In"
	NodeSelectorOpNotIn        NodeSelectorOperator = "NotIn"
	NodeSelectorOpExists       NodeSelectorOperator = "Exists"
	NodeSelectorOpDoesNotExist NodeSelectorOperator = "DoesNotExist"
	NodeSelectorOpGt           NodeSelectorOperator = "Gt"
	NodeSelectorOpLt           NodeSelectorOperator = "Lt"
)

type TopologySelectorTerm struct {
	MatchLabelExpressions []TopologySelectorLabelRequirement `json:"matchLabelExpressions,omitempty" protobuf:"bytes,1,rep,name=matchLabelExpressions"`
}

type TopologySelectorLabelRequirement struct {
	Key    string   `json:"key" protobuf:"bytes,1,opt,name=key"`
	Values []string `json:"values" protobuf:"bytes,2,rep,name=values"`
}
