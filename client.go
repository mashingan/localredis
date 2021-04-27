package localredis

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	internalStorage   sync.Map
	simpleStringRegex = regexp.MustCompile(`\+.*\r\n`)
	errorRegex        = regexp.MustCompile(`-.*\r\n`)
	integerRegex      = regexp.MustCompile(`:\d+\r\n`)
	bulkStringRegex   = regexp.MustCompile(`\$\d+\r\n`)
	arrayRegex        = regexp.MustCompile(`\*\d+\r\n`)
	defaultClient     = Client{}
)

type redisType byte
type Client struct {
	listener net.Listener
}

const (
	simpleStringType redisType = '+'
	errorType        redisType = '-'
	integerType      redisType = ':'
	bulkStringType   redisType = '$'
	arrayType        redisType = '*'
	terminal                   = "\r\n"
	bufferLength               = 4096
)

func fetchArray(inputbytes []byte) (values []interface{}, pos int, err error) {
	if len(inputbytes) > 2 && string(inputbytes[1:3]) == "-1" {
		pos = 5
		return
	}
	loc := arrayRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return
	}
	elemnum, converr := strconv.Atoi(string(inputbytes[loc[0]+1 : loc[1]-2]))
	if converr != nil {
		err = converr
		return
	}
	values = make([]interface{}, elemnum)
	pos = loc[1]
	rest := inputbytes[pos:]
	for i := 0; i < elemnum; i++ {
		if len(rest) <= 0 {
			break
		}
		switch redisType(rest[0]) {
		case arrayType:
			theval, newpos, theerr := fetchArray(rest)
			if newpos > 0 {
				pos += newpos
			}
			if newpos == 5 && len(theval) == 0 {
				values[i] = nil
			} else {
				values[i] = theval
			}
			err = theerr
			rest = rest[newpos:]
		case simpleStringType:
			str, newpos, theerr := fetchSimpleString(rest)
			if newpos > 0 {
				pos += newpos
			}
			values[i] = str
			err = theerr
			rest = rest[newpos:]
		case bulkStringType:
			str, newpos, theerr := fetchBulkString(rest)
			if newpos > 0 {
				pos += newpos
			}
			if str == "" && newpos == 5 {
				values[i] = nil
			} else {
				values[i] = str
			}
			err = theerr
			rest = rest[newpos:]
		case integerType:
			num, newpos, theerr := fetchInteger(rest)
			if newpos > 0 {
				pos += newpos
			}
			values[i] = num
			err = theerr
			rest = rest[newpos:]
		case errorType:
			valerr, newpos, theerr := fetchError(rest)
			if newpos > 0 {
				pos += newpos
			}
			values[i] = valerr
			err = theerr
			rest = rest[newpos:]
		}
	}
	return
}

func fetchSimpleString(inputbytes []byte) (string, int, error) {
	loc := simpleStringRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return "", 0, nil
	}
	return string(inputbytes[loc[0]+1 : loc[1]-2]), loc[1], nil
}

func fetchInteger(inputbytes []byte) (value int, pos int, err error) {
	loc := integerRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return
	}
	pos = loc[1]
	value, err = strconv.Atoi(string(inputbytes[loc[0]+1 : loc[1]-2]))
	return
}

func fetchBulkString(inputbytes []byte) (str string, pos int, err error) {
	if len(inputbytes) > 2 && string(inputbytes[1:3]) == "-1" {
		return "", 5, nil
	}
	loc := bulkStringRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return
	}
	num, converr := strconv.Atoi(string(inputbytes[loc[0]+1 : loc[1]-2]))
	err = converr
	if err != nil {
		return
	}
	pos = loc[1] + num + 2
	str = string(inputbytes[loc[1] : pos-2])
	return
}

func fetchError(inputbytes []byte) (errstr error, pos int, err error) {
	loc := errorRegex.FindIndex(inputbytes)
	if len(loc) == 0 {
		return
	}
	return fmt.Errorf("%s", string(inputbytes[loc[0]+1:loc[1]-2])), loc[1], nil
}

func ListenAndServe(addressPort string) error {
	l, err := net.Listen("tcp", addressPort)
	defaultClient.listener = l
	if err != nil {
		return err
	}
	defer func() {
		if l != nil {
			l.Close()
		}
	}()
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
	defer func() {
		if c != nil {
			c.Close()
		}
	}()
	var prevbuf []byte
	for {
		buff := make([]byte, bufferLength)
		if c == nil {
			return
		}
		n, err := c.Read(buff)
		if errors.Is(err, io.EOF) || n <= 0 {
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		var rest []byte
		if len(prevbuf) > 0 {
			rest = append(prevbuf, buff[:n]...)
		} else {
			rest = buff[:n]
		}
		_, rest, err = interpret(c, rest)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("rest:", rest)
		if len(rest) > 0 {
			prevbuf = append(prevbuf, rest...)
		} else {
			log.Println("rest is empty")
		}
	}
}

func interpret(c net.Conn, buff []byte) (complete bool, restbuf []byte, err error) {
	restbuf = buff
	switch redisType(buff[0]) {
	case simpleStringType:
		str, idx, errstr := fetchSimpleString(buff)
		restbuf = buff[idx:]
		log.Println(errstr)
		log.Println(str)
		err = errstr
	case integerType:
		valint, idx, errstr := fetchInteger(buff)
		restbuf = buff[idx:]
		log.Println(errstr)
		log.Println(valint)
		err = errstr
	case bulkStringType:
		str, idx, errstr := fetchBulkString(buff)
		restbuf = buff[idx:]
		log.Println(str)
		err = errstr
	case arrayType:
		vals, idx, errstr := fetchArray(buff)
		restbuf = buff[idx:]
		log.Println(vals)
		err = errstr
	case errorType:
		if len(buff) > 1 && buff[1] == '1' {
			return
		}
		valerr, idx, errstr := fetchError(buff)
		restbuf = buff[idx:]
		log.Println(valerr)
		err = errstr

	default:
	}
	return
}

func createNumRepr(n int) string {
	return fmt.Sprintf(":%d\r\n", n)
}

func createBulkString(input string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(input), input)
}

func createSimpleString(input string) string {
	return fmt.Sprintf("+%s\r\n", input)
}

func createArrayRepr(arrs []interface{}) (result string) {
	result = fmt.Sprintf("*%d\r\n", len(arrs))
	for _, ar := range arrs {
		switch varr := ar.(type) {
		case int:
			result += createNumRepr(varr)
		case string:
			if strings.Contains(varr, "\n") || strings.Contains(varr, "\x00") {
				result += createBulkString(varr)
			} else {
				result += createSimpleString(varr)
			}
		case []interface{}:
			result += createArrayRepr(varr)
		}
	}
	return
}

func Close() error {
	return defaultClient.listener.Close()
}
