package toyCache

import pb "github.com/toyCache/toyCache/toycachepb"

// PeerGetter is an interface must be implemented by a peer.
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}

// PeerPicker is an interface must be implemented to locate the peer
// that special key
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
