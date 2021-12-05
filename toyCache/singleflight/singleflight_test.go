package singleflight

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group
	v, err := g.Do("key", func() (interface{}, error) {
		return "foo", nil
	})
	got := fmt.Sprintf("%v (%T)", v, v)
	want := "foo (string)"
	require.Equal(t, want, got)
	require.NoError(t, err)
}

func TestDoDupSuppress(t *testing.T) {
	var g Group
	c := make(chan string)
	var calls int32
	fn := func() (interface{}, error){
		atomic.AddInt32(&calls, 1)
		return <-c, nil
	}
	var wg sync.WaitGroup
	for i:=0; i < 10; i++ {
		wg.Add(1)
		go func() {
			v, err := g.Do("key", fn)
			if err != nil{
				t.Errorf("Do error %v", err)
			}
			if v.(string) != "foo" {
				t.Errorf("got %v but expecte %v", v.(string), "foo")
			}
			wg.Done()
		}()
	}
	time.Sleep(time.Second)
	c <- "foo"
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("number of call expect 1 but got %v", got)
	}
}