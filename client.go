package localredis

import (
	"fmt"
	"net"
	"strings"
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

func ListenAndServe(addressPort string) error {
	l, err := net.Listen("tcp", addressPort)
	if err != nil {
		return err
	}
	defer l.Close()
	acceptingFailure := 0
	errorStackTrace := []error{}
	for {
		c, err := l.Accept()
		if err != nil {
			if acceptingFailure < 10 {
				errorStackTrace = append(errorStackTrace, err)
			} else {
				msgString := make([]string, len(errorStackTrace)+1)
				for i, preverr := range errorStackTrace {
					msgString[i] = preverr.Error()
				}
				msgString[len(errorStackTrace)] = err.Error()
				return fmt.Errorf(strings.Join(msgString, "\n"))
			}
			acceptingFailure++
		}
		go func(conn net.Conn) {

		}(c)
	}
}
