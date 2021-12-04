package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash map bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash     Hash
	replicas int
	keys     []int	//be sorted
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash: fn,
		replicas: replicas,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty return true if there is no item available
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// Add adds some keys to the hash
func (m *Map) Add(keys  ...string) {
	for _, key := range keys {
		for i:= 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets closet item in the hash to the provided key
func (m *Map) Get(key string) string{
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hash([]byte(key)))

	// Binary search appropriate replica
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx % len(m.keys)]]
}