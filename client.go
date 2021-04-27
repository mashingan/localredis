package localredis

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
)

var (
	internalStorage   sync.Map
	simpleStringRegex = regexp.MustCompile(`\+.*\s{2}`)
)

type redisType byte

const (
	simpleStringType redisType = '+'
	errorType        redisType = '-'
	integerType      redisType = ':'
	bulkStringType   redisType = '$'
	arrayType        redisType = '*'
	terminal                   = "\r\n"
	bufferLength               = 1024
)

func fetchArray(length int, inputbytes []byte) ([]interface{}, error) {
	return nil, fmt.Errorf("need implementation")
}

func fetchSimpleString(inputbytes []byte) (string, int, error) {
	loc := simpleStringRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return "", 0, nil
	}
	return string(inputbytes[loc[0]+1 : loc[1]-2]), loc[1], nil
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
		go handleCommand(c)
	}
}

func handleCommand(c net.Conn) {
	defer c.Close()
	for {
		buff := make([]byte, bufferLength)
		n, err := c.Read(buff)
		if errors.Is(err, io.EOF) || n <= 0 {
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		complete := false
		rest := buff
		for !complete {
			var err error
			complete, rest, err = interpret(c, rest)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func interpret(c net.Conn, buff []byte) (complete bool, restbuf []byte, err error) {
	switch redisType(buff[0]) {
	case simpleStringType:
		str, idx, errstr := fetchSimpleString(buff)
		restbuf = buff[idx:]
		log.Println(errstr)
		log.Println(str)
		err = errstr
	case integerType:
	case bulkStringType:
	case arrayType:
	default:
	}
	return
}
