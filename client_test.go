package localredis

import (
	"fmt"
	"strings"
	"testing"
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
}

func createBulkString(input string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
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

func createNumRepr(n int) string {
	return fmt.Sprintf(":%d\r\n", n)
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
