package main

import (
	localredis "local-redis"
)

func main() {
	localredis.ListenAndServe(redisListenAddr)
}

const (
	redisListenAddr = ":8099"
)
