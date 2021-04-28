package main

import "localredis"

func main() {
	localredis.ListenAndServe(redisListenAddr)
}

const (
	redisListenAddr = ":8099"
)
