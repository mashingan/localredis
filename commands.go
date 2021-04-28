package localredis

import (
	"fmt"
	"log"
	"net"
	"strings"
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
	"set":  setmap,
	"get":  getmap,
	"ping": pong,
	"quit": quit,
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
	if defaultClient.listener != nil {
		defaultClient.listener.Close()
	}
	sendOk(c)
}
