//
//
package node

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

// topologies key for bloom filter
func GetZoneKey(node *v1.Node) string {
	return node.Namespace
}
