package main

import (
	"flag"

	"github.com/mashingan/localredis"
)

var raddr = flag.String("addr", redisListenAddr, "set address to listen")

func main() {
	flag.Parse()
	localredis.ListenAndServe(*raddr)
}

const (
	redisListenAddr = ":8099"
)
