package toyCache

import (
	"errors"
	"github.com/toyCache/toyCache/singleflight"
	pb "github.com/toyCache/toyCache/toycachepb"
	"log"
	"sync"
)

// Group is a cache namespace and associate data load
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker

	// loadGroup make sure that each key fetched once
	// either in locally or remote
	loadGroup *singleflight.Group
}

// Getter load data for a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc implements Getter with a function
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create an instance of Group
func NewGroup(name string, cacheByte int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheByte},
		loadGroup: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup return a Group that create previously with NewGroup or return nil if there is no such group in groups
func GetGroup(name string) *Group {
	mu.RLock()
	group := groups[name]
	mu.RUnlock()
	return group
}

// RegisterPeer register a PeerPick for choosing a remote peer
func (g *Group) RegisterPeer(picker PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeer called more than once")
	}
	g.peers = picker
}

// Get return value for a key in cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("require key")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	// call Getter
	return g.load(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string)(ByteView, error) {
	req := &pb.Request{Group: g.name, Key: key}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// each key only fetched once regardless of the number of concurrent caller
	view, err := g.loadGroup.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok{
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[toyCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

