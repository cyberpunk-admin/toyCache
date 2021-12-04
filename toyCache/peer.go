package toyCache


// PeerGetter is an interface must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// PeerPicker is an interface must be implemented to locate the peer
// that special key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
