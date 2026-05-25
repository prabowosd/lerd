//go:build !linux

package dns

// LinkChanges is a no-op on non-Linux platforms. macOS containers receive
// DNS from the podman machine VM, so a host-side interface event would
// not change container resolution anyway, and the safety-net poll in
// each watcher covers any rare case we still want to react to.
func LinkChanges(out chan<- struct{}, done <-chan struct{}) {
	<-done
}
