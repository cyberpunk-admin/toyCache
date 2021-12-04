package toyCache

import (
	"fmt"
	"github.com/toyCache/toyCache/consistenthash"
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
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(view.ByteSlice())
	_, err = w.Write([]byte("\n"))
	if err != nil {
		return
	}
}

var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	baseURL string
}

func (g *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v",
		g.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
		)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v",res.StatusCode)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)