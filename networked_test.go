// +build local

package localredis

import (
	"errors"
	"io"
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

func TestGetexSetLocal(t *testing.T) {
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
	sethello := []interface{}{
		"set", "hello", "異世界",
	}
	nwrite, err := conn.Write([]byte(createReply(sethello)))
	if err != nil {
		t.Fatal(err)
	}
	if nwrite <= 0 {
		t.Error("failed to write, sent zero bytes")
	}
	buff := make([]byte, 128)
	nread, err := conn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if string(buff[:nread]) != "+OK\r\n" {
		t.Errorf("invalid reply, expected OK, got %s\n", buff[:nread])
	}

	getarg := createReply([]interface{}{"get", "hello"})
	t.Log(getarg)
	nwrite, err = conn.Write([]byte(getarg))
	if err != nil {
		t.Fatal(err)
	}
	if nwrite <= 0 {
		t.Error("failed to write, sent zero bytes")
	}
	nread, err = conn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	}

	getarg = createReply([]interface{}{
		"getex", "hello", "ex", 1,
	})
	t.Log(getarg)
	nwrite, err = conn.Write([]byte(getarg))
	if err != nil {
		t.Fatal(err)
	}
	if nwrite <= 0 {
		t.Error("failed to write, sent zero bytes")
	}
	nread, err = conn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	}
	time.Sleep(1 * time.Second)
	getarg = createReply([]interface{}{
		"getex", "hello", "ex", 1,
	})
	t.Log(getarg)
	nwrite, err = conn.Write([]byte(getarg))
	if err != nil {
		t.Fatal(err)
	}
	if nwrite <= 0 {
		t.Error("failed to write, sent zero bytes")
	}
	nread, err = conn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != "-1\r\n" {
		t.Errorf("invalid reply, expected error(-1), got %s\n", buff[:nread])
	}

	Close()
	w.Wait()
}
