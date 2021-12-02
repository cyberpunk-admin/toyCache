package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int{
	return len(d)
}

func TestCache_Get(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("123"))
	if v, ok := lru.Get("key1"); !ok || v.Len() != 3 {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestCache_Add(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key", String("123"))
	lru.Add("key", String("1234"))
	if lru.nBytes != int64(len("key") + len("1234")) {
		t.Fatal("expected 7 but got",lru.nBytes )
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	caps := len(k1 + v1 + k2 + v2)
	lru := New(int64(caps), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatalf("removeoldest item key1=value1 failed")
	}
}

func TestCache_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	OnEvicted := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), OnEvicted)

	lru.Add("key1", String("12345"))
	lru.Add("key2", String("12345"))
	lru.Add("key3", String("12345"))
	expect := []string{"key1", "key2"}

	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("Called OnEvicted failed, expect keys %s, but got %s", expect, keys)
	}
}