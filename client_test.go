package localredis

import (
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

func TestGetexSet(t *testing.T) {
	mconn := NewConnOverride()
	sethello := []interface{}{
		"set", "hello", "異世界",
	}
	setmap(mconn, sethello[1:])
	buff := make([]byte, 128)
	nread, err := mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	if string(buff[:nread]) != "+OK\r\n" {
		t.Errorf("invalid reply, expected OK, got %s\n", buff[:nread])
	}

	mconn.buffer.Reset()
	getarg := []interface{}{"get", "hello"}
	getmap(mconn, getarg[1:])
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	}

	mconn.buffer.Reset()
	getarg = []interface{}{
		"getex", "hello", "ex", 1,
	}
	getex(mconn, getarg[1:])
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected 異世界, got %s\n", buff[:nread])
	}
	time.Sleep(1 * time.Second)
	getarg = []interface{}{
		"get", "hello",
	}
	getmap(mconn, getarg[1:])
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	t.Log("buff:", string(buff))
	if string(buff[:nread]) != "-1\r\n" {
		t.Errorf("invalid reply, expected error(-1), got %s\n", buff[:nread])
	}

	mconn.Close()

}

func TestPersist(t *testing.T) {
	mconn := NewConnOverride()
	setarg := []interface{}{"hello", "異世界"}
	setmap(mconn, setarg)
	getarg := []interface{}{"hello"}
	getargEx := append(getarg, "px", 500)
	getex(mconn, getargEx)
	mconn.buffer.Reset()
	persist(mconn, getarg)
	buff := make([]byte, 128)
	nread, err := mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	if string(buff[:nread]) != ":1\r\n" {
		t.Errorf("invalid reply, expected 1, got %s\n", buff[:nread])
	}
	time.Sleep(500 * time.Millisecond)
	mconn.buffer.Reset()
	buff = make([]byte, 128)
	getmap(mconn, getarg)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	if string(buff[:nread]) != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected +異世界\\r\\n, got %s\n", buff[:nread])
	}
}

func TestTTL(t *testing.T) {
	mconn := NewConnOverride()
	setarg := []interface{}{"hello", "異世界"}
	setmap(mconn, setarg)
	buff := make([]byte, 64)
	nread, err := mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread := string(buff[:nread])
	if buffread != "+OK\r\n" {
		t.Errorf("invalid reply, expected OK, got %s\n", buffread)
	}
	getarg := []interface{}{"hello"}
	getargEx := append(getarg, "ex", 10)
	t.Log("getargEx:", getargEx)
	mconn.buffer.Reset()
	getex(mconn, getargEx)
	buff = make([]byte, 64)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread = string(buff[:nread])
	if buffread != createSimpleString("異世界") {
		t.Errorf("invalid reply, expected 異世界, got %s\n", buffread)
	}
	mconn.buffer.Reset()
	ttl(mconn, getarg)
	buff = make([]byte, 64)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread = string(buff[:nread])
	if !(buffread == ":10\r\n" || buffread == ":9\r\n") {
		t.Errorf("invalid reply, expected 9 or 10, got %s\n", buffread)
	}

	arg := []interface{}{"hello-2"}
	argex := append(arg, "px", 500)
	newarg := []interface{}{"hello-2", "新たな稼働"}
	setmap(mconn, newarg)
	buff = make([]byte, 64)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread = string(buff[:nread])
	if buffread != "+OK\r\n" {
		t.Errorf("invalid reply, expected OK, got %s\n", buffread)
	}
	getex(mconn, argex)
	buff = make([]byte, 64)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread = string(buff[:nread])
	if buffread != createSimpleString("新たな稼働") {
		t.Errorf("invalid reply, expected 新たな稼働, got %s\n", buffread)
	}
	mconn.buffer.Reset()
	pttl(mconn, arg)
	buff = make([]byte, 64)
	nread, err = mconn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if nread <= 0 {
		t.Error("could not read")
	}
	buffread = string(buff[:nread])
	var buffnum int
	fmt.Sscanf(buffread, ":%d\r\n", &buffnum)
	t.Log("buffread:", buffread)
	t.Log("buffnum:", buffnum)
	if buffnum != 500 {
		t.Errorf("invalid reply, expected ~500, got %d\n", buffnum)
	}
}

func TestExistKeys(t *testing.T) {
	conn := NewConnOverride()
	keys := []interface{}{"key1", "key2", "key3"}
	okreply := createSimpleString("OK")
	for i, k := range keys {
		setmap(conn, []interface{}{k, i + 1})
		buff := make([]byte, 10)
		n, err := conn.Read(buff)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
			if string(buff[:n]) != okreply {
				t.Errorf("invalid set operation, expect ok but got %s\n", buff[:n])
			}
		}
	}
	toCheckKeys := append(keys, "not-exists")
	args := []interface{}{"exists"}
	args = append(args, toCheckKeys...)
	conn.buffer.Reset()
	t.Log("args:", args)
	existsKeys(conn, args)
	buff := make([]byte, 32)
	nread, err := conn.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	buffread := string(buff[:nread])
	var buffnum int
	fmt.Sscanf(buffread, ":%d\r\n", &buffnum)
	if buffnum != 3 {
		t.Errorf("invalid checking keys, expect 3 keys, but got %d\n", buffnum)
	}
}

func TestCommandOverriding(t *testing.T) {
	conn := NewConnOverride()
	var (
		valueSet interface{}
		key      string
	)
	CommandOverride("set", func(c net.Conn, args []interface{}) {
		// set need minimum of 2 args
		if len(args) < 2 {
			SendError(c, fmt.Sprintf("invalid args, need minimum 2, got %d", len(args)))
			return
		}
		argkey, ok := args[0].(string)
		if !ok {
			SendError(c, fmt.Sprintf("invalid key type, need string, provided %T", args[0]))
			return
		}
		key = argkey
		valueSet = args[1]
		SendOk(c)
	})
	interpret(conn, []byte(CreateReply([]interface{}{
		"set", "hello", "異世界",
	})))
	buff := make([]byte, 32)
	nread, err := conn.Read(buff)
	if !(err == nil || errors.Is(err, io.EOF)) {
		t.Fatal(err)
	}
	buffread := string(buff[:nread])
	if buffread != createSimpleString("OK") {
		t.Errorf("fail reading reply, expecting +OK\\r\\n, got %s\n", buffread)
	}
	if key != "hello" {
		t.Fatalf("set override failed, expect key 'hello', got '%s'\n", key)
	}
	valueStr, ok := valueSet.(string)
	if !ok {
		t.Fatalf("invalid set value type, expect string, got %T\n", valueSet)
	}
	if valueStr != "異世界" {
		t.Fatalf("set override failed, expect value '異世界', got '%s'\n", valueStr)

	}
}
