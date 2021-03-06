package localredis

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var commandMap = map[string]CommandExecutioner{
	"set":     setmap,
	"get":     getmap,
	"ping":    pong,
	"quit":    quit,
	"getex":   getex,
	"persist": persist,
	"ttl":     ttl,
	"pptl":    pttl,
	"exists":  existsKeys,
}

type CommandExecutioner func(net.Conn, []interface{})

func SendError(c net.Conn, msg string) (int, error) {
	return c.Write([]byte(fmt.Sprintf("-%s\r\n", msg)))
}

func SendNil(c net.Conn) (int, error) {
	return c.Write([]byte("-1\r\n"))
}

func SendOk(c net.Conn) (int, error) {
	return c.Write([]byte("+OK\r\n"))
}

func SendValue(c net.Conn, value interface{}) (int, error) {
	return c.Write([]byte(CreateReply(value)))

}

func runCommand(c net.Conn, vals []interface{}) {
	if len(vals) < 1 {
		SendError(c, "invalid command format")
		return
	}
	command, ok := vals[0].(string)
	if !ok {
		SendError(c, "invalid command type")
		return
	}
	cmd, ok := commandMap[strings.ToLower(command)]
	if !ok {
		log.Printf("no command strings.ToLower(%s)\n", command)
		return
	}
	cmd(c, vals[1:])
}

func setmap(c net.Conn, args []interface{}) {
	if len(args) < 2 {
		SendError(c, fmt.Sprintf("invalid set command, need minimum 2 args, sent %d arg", len(args)))
		return
	}
	v, ok := args[0].(string)
	if !ok {
		SendError(c, fmt.Sprintf("invalid key format, expected string got %T", args[0]))
		return
	}
	if len(args[1:]) > 1 {
		setExpiration(c, args)
	}
	defaultClient.storage.Store(v, args[1])
	SendOk(c)
}

func getmap(c net.Conn, args []interface{}) {
	if len(args) < 1 {
		SendError(c, "invalid set command, need minimum 1 args, sent 0 arg")
	}
	switch v := args[0].(type) {
	case string:
		val, ok := defaultClient.storage.Load(v)
		if !ok {
			SendNil(c)
		}
		c.Write([]byte(CreateReply(val)))
		return
	}
	SendNil(c)
}

func pong(c net.Conn, args []interface{}) {
	c.Write([]byte(createSimpleString("PONG")))
}

func quit(c net.Conn, args []interface{}) {
	SendOk(c)
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
		SendNil(c)
		return
	}
	if len(rest) > 1 {
		setExpiration(c, args)
	}
	SendValue(c, val)
}

func persist(c net.Conn, args []interface{}) {
	if len(args) < 1 {
		SendError(c, "invalid format, no key sent")
		return
	}
	key, ok := args[0].(string)
	if !ok {
		SendError(c, "invalid key type, need string")
		return
	}
	_, avail := defaultClient.storage.Load(key)
	_, persisted := defaultClient.persist[key]
	if avail && !persisted {
		defaultClient.persist[key] = true
		SendValue(c, 1)
		return
	}
	if !avail && !persisted {
		SendValue(c, 0)
		return
	}
	SendValue(c, 0)

}

func durationCalc(num int, timesetter string) (dur time.Duration) {
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
	return
}

func timeout(arg interface{}, dur time.Duration) {
	time.Sleep(dur)
	keystr := arg.(string)
	persist, ok := defaultClient.persist[keystr]
	timeout, hasTo := defaultClient.timeout[keystr]
	if !ok {
		defaultClient.storage.Delete(arg)
		delete(defaultClient.persist, keystr)
		delete(defaultClient.timeout, keystr)
	} else if persist || (hasTo && timeout.Before(time.Now())) {
		delete(defaultClient.persist, keystr)
		defaultClient.timeout[keystr] = time.Unix(0, 0)

	}
}

func setExpiration(c net.Conn, args []interface{}) {
	key := args[0]
	rest := args[1:]
	timesetter, ok := rest[0].(string)
	if !ok {
		SendError(c, "invalid expiration option")
		return
	}
	timesetter = strings.ToLower(timesetter)
	if !validopt(timesetter) {
		SendError(c, fmt.Sprintf(
			"invalid expiration option, sent %s expected one of ex, px, eaxt, pxat", timesetter))
		return
	}
	num, ok := rest[1].(int)
	if !ok {
		SendError(c, fmt.Sprintf("invalid numeric expiration, got %#v", rest[1]))
		return
	}
	dur := durationCalc(num, timesetter)
	log.Println("dur:", dur)
	keystr := key.(string)
	defaultClient.timeout[keystr] = time.Now().Add(dur)
	go timeout(key, dur)
}

type ttlKind string

const (
	ttlSecond      ttlKind = "second"
	ttlMillisecond ttlKind = "millisecond"
)

func ttlimp(c net.Conn, args []interface{}, kind ttlKind) {
	if len(args) < 1 {
		SendError(c, "invalid format, no key sent")
		return
	}
	key, ok := args[0].(string)
	if !ok {
		SendError(c, "invalid key type, need string")
		return
	}
	_, avail := defaultClient.storage.Load(key)
	_, persisted := defaultClient.persist[key]
	until, hasTimeout := defaultClient.timeout[key]
	if !avail {
		SendNil(c)
		return
	}
	if !persisted && !hasTimeout {
		SendValue(c, -1)
		return
	}
	if until.Before(time.Now()) {
		SendValue(c, -1)
		return
	}
	var secondToLive int
	switch kind {
	case ttlSecond:
		secondToLive = int(time.Until(until).Round(time.Second).Seconds())
	case ttlMillisecond:
		secondToLive = int(time.Until(until).Round(time.Millisecond).Milliseconds())
	}
	SendValue(c, int(secondToLive))
}

func ttl(c net.Conn, args []interface{}) {
	ttlimp(c, args, "second")
}

func pttl(c net.Conn, args []interface{}) {
	ttlimp(c, args, "millisecond")
}

func existsKeys(c net.Conn, args []interface{}) {
	if len(args) < 1 {
		SendValue(c, 0)
		return
	}
	totalKeys := 0
	for _, key := range args {
		_, ok := defaultClient.storage.Load(key)
		if ok {
			totalKeys++
		}
	}
	SendValue(c, totalKeys)
}
