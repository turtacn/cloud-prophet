//
package cache

import (
	"fmt"

	utilnode "github.com/turtacn/cloud-prophet/scheduler/helper/node"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/klog/v2"
)

type nodeTree struct {
	tree      map[string]*nodeArray // a map from zone (region-zone) to an array of nodes in the zone.
	zones     []string              // a list of all the zones in the tree (keys)
	zoneIndex int
	numNodes  int
}

type nodeArray struct {
	nodes     []string
	lastIndex int
}

func (na *nodeArray) next() (nodeName string, exhausted bool) {
	if len(na.nodes) == 0 {
		klog.Error("The nodeArray is empty. It should have been deleted from NodeTree.")
		return "", false
	}
	if na.lastIndex >= len(na.nodes) {
		return "", true
	}
	nodeName = na.nodes[na.lastIndex]
	na.lastIndex++
	return nodeName, false
}

func newNodeTree(nodes []*v1.Node) *nodeTree {
	nt := &nodeTree{
		tree: make(map[string]*nodeArray),
	}
	for _, n := range nodes {
		nt.addNode(n)
	}
	return nt
}

func (nt *nodeTree) addNode(n *v1.Node) {
	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for _, nodeName := range na.nodes {
			if nodeName == n.Name {
				klog.Warningf("node %q already exist in the NodeTree", n.Name)
				return
			}
		}
		na.nodes = append(na.nodes, n.Name)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = &nodeArray{nodes: []string{n.Name}, lastIndex: 0}
	}
	klog.Infof("Added node %q in group %q to NodeTree", n.Name, zone)
	nt.numNodes++
}

func (nt *nodeTree) removeNode(n *v1.Node) error {
	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, nodeName := range na.nodes {
			if nodeName == n.Name {
				na.nodes = append(na.nodes[:i], na.nodes[i+1:]...)
				if len(na.nodes) == 0 {
					nt.removeZone(zone)
				}
				klog.Infof("Removed node %q in group %q from NodeTree", n.Name, zone)
				nt.numNodes--
				return nil
			}
		}
	}
	klog.Errorf("Node %q in group %q was not found", n.Name, zone)
	return fmt.Errorf("node %q in group %q was not found", n.Name, zone)
}

func (nt *nodeTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}

func (nt *nodeTree) updateNode(old, new *v1.Node) {
	var oldZone string
	if old != nil {
		oldZone = utilnode.GetZoneKey(old)
	}
	newZone := utilnode.GetZoneKey(new)
	if oldZone == newZone {
		return
	}
	nt.removeNode(old) // No error checking. We ignore whether the old node exists or not.
	nt.addNode(new)
}

func (nt *nodeTree) resetExhausted() {
	for _, na := range nt.tree {
		na.lastIndex = 0
	}
	nt.zoneIndex = 0
}

func (nt *nodeTree) next() string {
	if len(nt.zones) == 0 {
		return ""
	}
	numExhaustedZones := 0
	for {
		if nt.zoneIndex >= len(nt.zones) {
			nt.zoneIndex = 0
		}
		zone := nt.zones[nt.zoneIndex]
		nt.zoneIndex++
		nodeName, exhausted := nt.tree[zone].next()
		if exhausted {
			numExhaustedZones++
			if numExhaustedZones >= len(nt.zones) { // all zones are exhausted. we should reset.
				nt.resetExhausted()
			}
		} else {
			return nodeName
		}
	}
}
