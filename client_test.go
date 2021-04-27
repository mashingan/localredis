package localredis

import (
	"testing"
)

func TestFetchSimpleString(t *testing.T) {
	// simple string cannot hold utf-8 byte codes and will fail
	// expected := "hello 異世界"
	// rawbyte := []byte(`\+hello 異世界\r\n`)
	expected := "hello world"
	rawbyte := []byte("+hello world\r\n")
	fetchstr, pos, _ := fetchSimpleString(rawbyte)
	t.Log("fetchstr:", fetchstr)
	t.Log("pos:", pos)
	if fetchstr != expected {
		t.Errorf("Invalid string fetch, expected %s, got %s\n", expected, fetchstr)
	}
	if pos != len(rawbyte) {
		t.Errorf("Invalid position, expected %d, got %d\n", len(rawbyte), pos)
	}
}
