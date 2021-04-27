package localredis

import (
	"fmt"
	"sync"
)

var (
	internalStorage sync.Map
)

type redisType byte

const (
	simpleStringType redisType = '+'
	errorType        redisType = '-'
	integerType      redisType = ':'
	bulkStringType   redisType = '$'
	arrayType        redisType = '*'
	terminal                   = "\r\n"
)

func fetchArray(length int, inputbytes []byte) ([]interface{}, error) {
	return nil, fmt.Errorf("need implementation")
}

func fetchSimpleString(inputbytes []byte) (string, error) {
	return "", fmt.Errorf("need implementation")
}

func fetchInteger(inputbytes []byte) (int, error) {
	return 0, fmt.Errorf("need implementation")
}

func fetchBulkString(strcount int, inputbytes []byte) ([]string, error) {
	return nil, fmt.Errorf("need implementation")
}
