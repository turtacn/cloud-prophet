//
package base

type ClientSet interface {
	PodsGetter
	NodesGetter
	AzGetter
}

type PodsGetter interface {
}
type NodesGetter interface {
}
type AzGetter interface {
}
