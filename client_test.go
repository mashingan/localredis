package localredis

import (
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
