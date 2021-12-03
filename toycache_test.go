package toyCache

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Bob":  "123",
	"Jack": "555",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	v, _ := f.Get("key")
	require.Equal(t, expect, v)
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	toyC := NewGroup("score", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[slowDB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key]++
			return []byte(v), nil
		}
		return nil, fmt.Errorf("key: %s not exist", key)
	}))

	for k, v := range db {
		view, err := toyC.Get(k)
		require.NoError(t, err)
		require.Equal(t, view.String(), v)
		_, err = toyC.Get(k)
		require.NoError(t, err)
		require.True(t, loadCounts[k] == 1, fmt.Errorf("cache %s miss", k))
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2 << 10, GetterFunc(func(key string) ([]byte, error) {
		return []byte{}, nil
	}))
	group := GetGroup(groupName)
	require.NotEmpty(t, group)
	require.Equal(t, groupName, group.name, fmt.Errorf("group %s not exist", groupName))
	group = GetGroup(groupName + "invalid")
	require.Empty(t, group)
}
