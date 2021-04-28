// +build local

package localredis

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	var w sync.WaitGroup
	w.Add(1)
	addr := ":9025"
	// go ListenAndServe(addr)
	go func(wg *sync.WaitGroup) {
		defer w.Done()
		ListenAndServe(addr)
	}(&w)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	orig := "hello world"
	raw := []byte(createSimpleString(orig))
	n, err := conn.Write(raw)
	t.Log("sent:", n)
	if err != nil {
		t.Error(err)
	}
	if n < 1 {
		t.Errorf("invalid sending, got sent 0, expected %d\n", len(raw))
	}

	orig = "hello 異世界"
	raw = []byte(createBulkString(orig))
	n, err = conn.Write(raw)
	t.Log("sent:", n)
	if err != nil {
		t.Error(err)
	}
	if n < 1 {
		t.Errorf("invalid sending, got sent 0, expected %d\n", len(raw))
	}
	time.Sleep(1 * time.Second)
	Close()
	w.Wait()
	// raw = []byte(createArrayRepr([]interface{}{
	// 	"hello world"
	// }))
}
