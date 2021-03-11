package node

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

func GetZoneKey(node *v1.Node) string {
	return node.Namespace
}
