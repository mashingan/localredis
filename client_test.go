package localredis

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func assertStr(t *testing.T, fetchstr, expected string, pos, expectpos int) {
	if fetchstr != expected {
		t.Errorf("Invalid string fetch, expected %s, got %s\n", expected, fetchstr)
	}
	if pos != expectpos {
		t.Errorf("Invalid position, expected %d, got %d\n", expectpos, pos)
	}

}

func TestFetchSimpleString(t *testing.T) {
	// simple string cannot hold utf-8 byte codes and will fail
	// expected := "hello 異世界"
	// rawbyte := []byte(`\+hello 異世界\r\n`)
	expected := "hello world"
	rawbyte := []byte("+hello world\r\n")
	fetchstr, pos, _ := fetchSimpleString(rawbyte)
	t.Log("fetchstr:", fetchstr)
	t.Log("pos:", pos)
	assertStr(t, fetchstr, expected, pos, len(rawbyte))

	expected = "hello    world"
	rawbyte = []byte("+hello    world\r\nnananan")
	fetchstr, pos, _ = fetchSimpleString(rawbyte)
	assertStr(t, fetchstr, expected, pos,
		strings.Index(string(rawbyte), expected)+len(expected)+2)

	expected = "++hello    world--"
	rawbyte = []byte("+++hello    world--\r\nnananan")
	fetchstr, pos, _ = fetchSimpleString(rawbyte)
	assertStr(t, fetchstr, expected, pos,
		strings.Index(string(rawbyte), expected)+len(expected)+2)

	expected = ""
	rawbyte = []byte("+++hello    world--nnananan")
	fetchstr, pos, _ = fetchSimpleString(rawbyte)
	assertStr(t, fetchstr, expected, pos, 0)
}

func testBulk(t *testing.T, s string) {
	orig := "hello of nice world"
	bulkstr := createBulkString(orig)
	fetchstr, pos, _ := fetchBulkString([]byte(bulkstr))
	assertStr(t, fetchstr, orig, pos, len(bulkstr))

}

func TestBulkString(t *testing.T) {
	testBulk(t, "hello of nice world")
	testBulk(t, "hello 異世界")
	orig := ""
	bulkstr := []byte("$-1\r\n")
	fetchstr, pos, _ := fetchBulkString(bulkstr)
	assertStr(t, fetchstr, orig, pos, len(bulkstr))
}

func assertNum(t *testing.T, got, expect, pos, expectpos int) {
	if got != expect {
		t.Errorf("Invalid string fetch, expect %d, got %d\n", expect, got)
	}
	if pos != expectpos {
		t.Errorf("Invalid position, expected %d, got %d\n", expectpos, pos)
	}

}

func testNum(t *testing.T, n int) {
	raw := createNumRepr(n)
	fetchnum, pos, _ := fetchInteger([]byte(raw))
	assertNum(t, fetchnum, n, pos, len(raw))
}

func TestInteger(t *testing.T) {
	testNum(t, 10)
	testNum(t, 2555)
}

func TestArray(t *testing.T) {
	rawstr := "*5\r\n"
	for i := 1; i < 5; i++ {
		rawstr += fmt.Sprintf(":%d\r\n", i)
	}
	rawstr += createBulkString("Foobar")
	actualraw := []byte("*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nFoobar\r\n")
	if rawstr != string(actualraw) {
		t.Fatalf("invalid creating array: got %s expected %s\n", rawstr, string(actualraw))
	}
	arrs, pos, _ := fetchArray(actualraw)
	if len(arrs) == 0 {
		t.Fatalf("empty values, expected %d elements", 5)
	}
	if pos != len(actualraw) {
		t.Errorf("invalid pos, got %d expected %d\n", pos, len(actualraw))
	}

	for i, o := range arrs[:3] {
		num, ok := o.(int)
		if !ok {
			t.Errorf("invalid integer num: %v\n", o)
			continue
		}
		if num != i+1 {
			t.Errorf("invalid integer value: got %d expected %d\n", num, i+1)
		}
	}
	foobar, ok := arrs[4].(string)
	if !ok {
		t.Fatalf("cannot convert/invalid %v to string\n", arrs[4])
	}
	if foobar != "Foobar" {
		t.Errorf("invalid string value, got '%s' expected Foobar\n", foobar)
	}

	// another case
	rawstr = "*2\r\n"
	rawstr += "*3\r\n"
	for i := 1; i <= 3; i++ {
		rawstr += fmt.Sprintf(":%d\r\n", i)
	}
	rawstr += "*2\r\n+Foo\r\n-Bar\r\n"
	actualraw = []byte("*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n")
	if rawstr != string(actualraw) {
		t.Fatalf("invalid raw, got %s expected %s\r\n", rawstr, string(actualraw))
	}
	arrs, pos, _ = fetchArray(actualraw)
	if len(arrs) == 0 {
		t.Fatalf("empty values, expected %d elements", 2)
	}
	if pos != len(actualraw) {
		t.Errorf("invalid pos, got %d expected %d\n", pos, len(actualraw))
	}
	arr1, ok := arrs[0].([]interface{})
	if !ok {
		t.Fatalf("invalid format, got %#v expected array\n", arrs[0])
	}
	if len(arr1) != 3 {
		t.Errorf("invalid length, got %d expected 3\n", len(arr1))
	}
	for i, o := range arr1 {
		num, ok := o.(int)
		if !ok {
			t.Errorf("invalid integer num: %v\n", o)
			continue
		}
		if num != i+1 {
			t.Errorf("invalid integer value: got %d expected %d\n", num, i+1)
		}
	}
	arr2, ok := arrs[1].([]interface{})
	if !ok {
		t.Fatalf("invalid format, got %#v expected array\n", arrs[1])
	}
	if len(arr1) != 3 {
		t.Errorf("invalid length, got %d expected 2\n", len(arr2))
	}
	foo, ok := arr2[0].(string)
	if !ok {
		t.Fatalf("invalid format, got %#v, expected string", arr2[0])
	}
	if foo != "Foo" {
		t.Errorf("invalid string value, got %s, expected Foo", foo)

	}
	theerr, ok := arr2[1].(error)
	if !ok {
		t.Fatalf("invalid format, got %#v, expected error", arr2[1])
	}
	if theerr == nil || theerr.Error() != "Bar" {
		t.Errorf("invalid error value, got %v, expected Bar", theerr)
	}

}

type mockConn struct {
	buffer *bytes.Buffer
	closed bool
	mockAddr
	deadline time.Time
}

func (m *mockConn) Write(p []byte) (int, error) {
	return m.buffer.Write(p)
}

func (m *mockConn) Read(p []byte) (int, error) {
	return m.buffer.Write(p)
}

func (m *mockConn) Close() error {
	if !m.closed {
		return nil
	}
	return fmt.Errorf("already closed")
}

type mockAddr struct{}

func (m *mockAddr) Network() string {
	return "tcp"
}
func (m *mockAddr) String() string {
	return "127.0.0.1"
}

func (m *mockConn) LocalAddr() net.Addr {
	return &m.mockAddr
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &m.mockAddr
}

func (m *mockConn) SetDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func newMockConn() *mockConn {
	m := new(mockConn)
	m.buffer = new(bytes.Buffer)
	return m
}

func TestGetexSet(t *testing.T) {
	// conn := newMockConn()
	mconn := newMockConn()
	sethello := []interface{}{
		"set", "hello", "異世界",
	}
	setmap(mconn, sethello[1:])
	buff := make([]byte, 128)
	nread, err := mconn.buffer.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	t.Log("buff:", buff)
	t.Log("nread:", nread)
	if string(buff[:nread]) != "+OK\r\n" {
		t.Errorf("invalid reply, expected OK, got %s\n", buff[:nread])
	}
	// nwrite, err := conn.Write([]byte(createReply(sethello)))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if nwrite <= 0 {
	// 	t.Error("failed to write, sent zero bytes")
	// }
	// buff := make([]byte, 128)
	// nread, err := conn.Read(buff)
	// if err != nil && !errors.Is(err, io.EOF) {
	// 	t.Fatal(err)
	// }
	// if string(buff[:nread]) != "+OK\r\n" {
	// 	t.Errorf("invalid reply, expected OK, got %s\n", buff[:nread])
	// }

	// getarg := createReply([]interface{}{"get", "hello"})
	// t.Log(getarg)
	// nwrite, err = conn.Write([]byte(getarg))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if nwrite <= 0 {
	// 	t.Error("failed to write, sent zero bytes")
	// }
	// nread, err = conn.Read(buff)
	// if err != nil && !errors.Is(err, io.EOF) {
	// 	t.Fatal(err)
	// }
	// t.Log("buff:", string(buff))
	// if string(buff[:nread]) != createSimpleString("異世界") {
	// 	t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	// }

	// getarg = createReply([]interface{}{
	// 	"getex", "hello", "ex", 1,
	// })
	// t.Log(getarg)
	// nwrite, err = conn.Write([]byte(getarg))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if nwrite <= 0 {
	// 	t.Error("failed to write, sent zero bytes")
	// }
	// nread, err = conn.Read(buff)
	// if err != nil && !errors.Is(err, io.EOF) {
	// 	t.Fatal(err)
	// }
	// t.Log("buff:", string(buff))
	// if string(buff[:nread]) != createSimpleString("異世界") {
	// 	t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	// }
	// time.Sleep(1 * time.Second)
	// getarg = createReply([]interface{}{
	// 	"getex", "hello", "ex", 1,
	// })
	// t.Log(getarg)
	// nwrite, err = conn.Write([]byte(getarg))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if nwrite <= 0 {
	// 	t.Error("failed to write, sent zero bytes")
	// }
	// nread, err = conn.Read(buff)
	// if err != nil && !errors.Is(err, io.EOF) {
	// 	t.Fatal(err)
	// }
	// t.Log("buff:", string(buff))
	// if string(buff[:nread]) != "-1\r\n" {
	// 	t.Errorf("invalid reply, expected error(-1), got %s\n", buff[:nread])
	// }

	// Close()

}
