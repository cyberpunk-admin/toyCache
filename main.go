package main

import (
	"flag"
	"fmt"
	"github.com/toyCache/toyCache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "532",
	"Bob":  "700",
}

func CreateGroup() *toyCache.Group {
	return toyCache.NewGroup("scores", 2 << 10, toyCache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, group *toyCache.Group) {
	peers := toyCache.NewHTTPPool(addr)
	peers.Set(addrs...)

	group.RegisterPeer(peers)
	log.Println("toyCache is running at", addr)
	// addr eg. http://example.com:8080
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *toyCache.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = w.Write(view.ByteSlice())
		if err != nil {
			return
		}
	}))
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "toyCache server port")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string {
		8001 : "http://localhost:8001",
		8002 : "http://localhost:8002",
		8003 : "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	group := CreateGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}
