package main

import (
	"fmt"
	"net"
)

func main() {}

const (
	//lint:ignore U1000 this will be used for 2nd case of TestRequestCalc
	redisListenAddr = ":8099"
)

//lint:ignore U1000 this will be used for 2nd case of TestRequestCalc
func errReply(c net.Conn, err error) error {
	_, _ = c.Write([]byte(fmt.Sprintf("-%s\r\n", err.Error())))
	return err
}

//lint:ignore U1000 this will be used for 2nd case of TestRequestCalc
type redisOp func(c net.Conn) error

//lint:ignore U1000 this will be used for 2nd case of TestRequestCalc
func okRedisOp(c net.Conn) error {
	_, err := c.Write([]byte("+OK\r\n"))
	return err
}

//lint:ignore U1000 this will be used for 2nd case of TestRequestCalc
// func setKey(c net.Conn) error {
// 	reset := regexp.MustCompile(`\$\d+\s{2}set`)
// 	re := regexp.MustCompile(`\$\d+\s{2}.*\s{2}`)
// 	recount := regexp.MustCompile(`\*\d+`)
// 	buf := make([]byte, 512)
// 	n, err := c.Read(buf)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	if n <= 0 {
// 		return nil
// 	}
// 	if !reset.Match(buf[:n]) {
// 		return nil
// 	}
// 	countbyte := recount.Find(buf[:n])
// 	if countbyte == nil {
// 		return nil
// 	}
// 	count, _ := strconv.Atoi(string(countbyte[1:]))
// 	log.Println("count:", count)
// 	allBytes := re.FindAll(buf[:n], -1)
// 	log.Printf("allBytes: %q\n", allBytes)
// 	key = string(allBytes[1])
// 	splits := strings.Split(key, "\r\n")
// 	key = splits[len(splits)-2]

// 	credstr := string(allBytes[2])
// 	splits = strings.Split(credstr, "\r\n")
// 	log.Println("splits:", splits)
// 	credload := []byte(splits[len(splits)-2])
// 	log.Println("credload:", string(credload))
// 	return json.Unmarshal(credload, &appstate)
// }

// func getKey(c net.Conn) error {
// 	reset := regexp.MustCompile(`\$\d+\s{2}get`)
// 	re := regexp.MustCompile(`\$\d+\s{2}.*\s{2}`)
// 	recount := regexp.MustCompile(`\*\d+`)
// 	buf := make([]byte, 512)
// 	n, err := c.Read(buf)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	if n <= 0 {
// 		return nil
// 	}
// 	if !reset.Match(buf[:n]) {
// 		return nil
// 	}
// 	countbyte := recount.Find(buf[:n])
// 	if countbyte == nil {
// 		return nil
// 	}
// 	log.Println(countbyte[1:])
// 	count, err := strconv.Atoi(string(countbyte[1:]))
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	allBytes := re.FindAll(buf[:n], -1)
// 	getkey := string(allBytes[count-1])
// 	splits := strings.Split(getkey, "\r\n")
// 	getkey = splits[len(splits)-2]
// 	log.Println("getkey:", getkey)
// 	if getkey == shared {
// 		_, err = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n",
// 			len(sharedValue), sharedValue)))
// 		return err
// 	}

// 	if getkey != key {
// 		err := fmt.Errorf("key '%s' is not available", getkey)
// 		errReply(c, err)
// 		return err
// 	}

// 	credjson, _ := json.Marshal(appstate)
// 	_, err = c.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(credjson), string(credjson))))

// 	return err
// }

// func redisDummy(w *sync.WaitGroup, oneshot bool, ops ...redisOp) {
// 	redisListenAddr := ":8099"
// 	w.Done()
// 	l, err := net.Listen("tcp", redisListenAddr)
// 	if err != nil {
// 		log.Fatal(err)
// 		return
// 	}
// 	defer l.Close()
// 	if oneshot {
// 		c, err := l.Accept()
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		defer c.Close()
// 		for _, op := range ops {
// 			if err := op(c); err != nil {
// 				log.Println(err)
// 				return
// 			}
// 		}

// 	} else {
// 		for {
// 			c, err := l.Accept()
// 			if err != nil {
// 				log.Println(err)
// 				return
// 			}
// 			go func(c net.Conn) {
// 				defer c.Close()
// 				for {
// 					count := 0
// 					opslen := len(ops)
// 					idx := count % opslen
// 					op := ops[idx]
// 					count++
// 					if err := op(c); err != nil {
// 						log.Println(err)
// 						return
// 					}
// 				}

// 			}(c)
// 		}
// 	}
// }
