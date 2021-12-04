package consistenthash

import (
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// Override the hash function to return simple value
	// the key can convent to integer
	hash := New(3, func(data []byte) uint32 {
		i, err := strconv.Atoi(string(data))
		if err != nil {
			panic(err)
		}
		return uint32(i)
	})
	// Given above hash function, this will give replicas as
	// 2 ,4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("2", "4", "6")

	//fmt.Println(hash.keys)
	testCase := map[string]string{
		"2":  "2",
		"3":  "4",
		"17": "2",
		"27": "2",
	}
	for k, v := range testCase {
		require.Equal(t, v, hash.Get(k))
	}
	hash.Add("8")
	testCase["17"] = "8"
	testCase["27"] = "8"
	for k, v := range testCase {
		require.Equal(t, v, hash.Get(k))
	}
}

func TestConsistency(t *testing.T) {
	hash1 := New(1, nil)
	hash2 := New(1, nil)

	hash1.Add("Jack", "Tom", "Bob")
	hash2.Add("Tom", "Jack", "Bob")
	require.Equal(t, hash1.Get("Bill"), hash2.Get("Bill"), "Fetch Bill form both hash should same")

	hash2.Add("Becky", "Ben", "Bobby")

	if hash1.Get("Ben") != hash2.Get("Ben") ||
		hash1.Get("Bob") != hash2.Get("Bob") ||
		hash1.Get("Bonny") != hash2.Get("Bonny") {
		t.Errorf("Direct matches should always return the same entry")
	}

}
