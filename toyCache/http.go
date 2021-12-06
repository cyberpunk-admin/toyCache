package toyCache

import (
	"fmt"
	"github.com/toyCache/toyCache/consistenthash"
	pb "github.com/toyCache/toyCache/toycachepb"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_toyCache"
	defaultReplicas = 50
)

// HTTPPool implement a PeerPick for a pool of HTTP peers.
type HTTPPool struct {
	self       string
	basePath   string
	mu         sync.Mutex // guards peer and httpGetter
	peers      *consistenthash.Map
	httpGetter map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

// NewHTTPPool initializes an HTTP pool of peers
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log HTTPPool info with peer name
func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v))
}

// Set update the pool' list of peer,
// each peer should be a valid URL
// for example http://example.net:8080
func (h *HTTPPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.peers = consistenthash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetter[peer] = &httpGetter{baseURL: peer + h.basePath}
	}
}

// PickPeer pick a peer according a key
func (h *HTTPPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.peers.IsEmpty() {
		return nil, false
	}
	if peer := h.peers.Get(key); peer != h.self {
		return h.httpGetter[peer], true
	}
	return nil, false
}

// ServeHTTP handle all http request
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	h.Log("%s %s", r.Method, r.URL.Path)

	// /<basePath>/<groupName>/<key> required
	//fmt.Printf("[urlPath] %s", r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(h.basePath)+1:], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := groups[groupName]
	if group == nil {
		http.Error(w, fmt.Sprintf("no such group: %s", groupName), http.StatusBadRequest)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the value to the response body as a proto message
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	baseURL string
}

func (g *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v",
		g.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
		)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %v",res.StatusCode)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return err
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)