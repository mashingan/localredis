package localredis

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type commandExecutioner func(net.Conn, []interface{})

func sendError(c net.Conn, msg string) (int, error) {
	return c.Write([]byte(fmt.Sprintf("-%s\r\n", msg)))
}

func sendNil(c net.Conn) (int, error) {
	return c.Write([]byte("-1\r\n"))
}

func sendOk(c net.Conn) (int, error) {
	return c.Write([]byte("+OK\r\n"))
}

func sendValue(c net.Conn, value interface{}) (int, error) {
	return c.Write([]byte(createReply(value)))

}

func runCommand(c net.Conn, vals []interface{}) {
	if len(vals) < 1 {
		sendError(c, "invalid command format")
		return
	}
	command, ok := vals[0].(string)
	if !ok {
		sendError(c, "invalid command type")
		return
	}
	cmd, ok := commandMap[strings.ToLower(command)]
	if !ok {
		log.Printf("no command strings.ToLower(%s)\n", command)
		return
	}
	cmd(c, vals[1:])
}

var commandMap = map[string]commandExecutioner{
	"set":   setmap,
	"get":   getmap,
	"ping":  pong,
	"quit":  quit,
	"getex": getex,
}

func setmap(c net.Conn, args []interface{}) {
	if len(args) < 2 {
		sendError(c, fmt.Sprintf("invalid set command, need minimum 2 args, sent %d arg", len(args)))
	}
	switch v := args[0].(type) {
	case string:
		defaultClient.storage.Store(v, args[1])
		sendOk(c)
	}
}

func getmap(c net.Conn, args []interface{}) {
	if len(args) < 1 {
		sendError(c, "invalid set command, need minimum 1 args, sent 0 arg")
	}
	switch v := args[0].(type) {
	case string:
		val, ok := defaultClient.storage.Load(v)
		if !ok {
			sendNil(c)
		}
		c.Write([]byte(createReply(val)))
	}
}

func pong(c net.Conn, args []interface{}) {
	c.Write([]byte(createSimpleString("PONG")))
}

func quit(c net.Conn, args []interface{}) {
	sendOk(c)
	if defaultClient.listener != nil {
		defaultClient.listener.Close()
	}
}

var expireSettingOpt = []string{"ex", "px", "exat", "pxat"}

func validopt(s string) bool {
	for _, o := range expireSettingOpt {
		if o == s {
			return true
		}
	}
	return false
}
func getex(c net.Conn, args []interface{}) {
	if len(args) <= 1 {
		getmap(c, args)
		return
	}
	rest := args[1:]
	key := args[0]
	val, ok := defaultClient.storage.Load(key)
	if !ok {
		sendNil(c)
		return
	}
	if len(rest) > 1 {
		timesetter, ok := rest[0].(string)
		if !ok {
			sendError(c, "invalid expiration option")
			return
		}
		timesetter = strings.ToLower(timesetter)
		if !validopt(timesetter) {
			sendError(c, fmt.Sprintf(
				"invalid expiration option, sent %s expected one of ex, px, eaxt, pxat", timesetter))
			return
		}
		num, ok := rest[1].(int)
		if !ok {
			sendError(c, "invalid numeric expiration")
			return
		}
		var dur time.Duration
		switch timesetter {
		case "ex":
			dur = time.Duration(num) * time.Second
		case "px":
			dur = time.Duration(num) * time.Millisecond
		case "exat":
			dur = time.Until(time.Unix(int64(num), 0))
		case "pxat":
			secnum := int64(num / 1000)
			milnum := (int64(num) % secnum) * 1e6
			dur = time.Until(time.Unix(secnum, milnum))
		}
		go func(arg interface{}, dur time.Duration) {
			time.Sleep(dur)
			keystr := arg.(string)
			persist, ok := defaultClient.persist[keystr]
			if !ok {
				defaultClient.storage.Delete(arg)
			} else if persist {
				delete(defaultClient.persist, keystr)
			}
		}(key, dur)
	}
	sendValue(c, val)
}
