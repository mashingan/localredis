package main

import "github.com/mashingan/localredis"

func main() {
	localredis.ListenAndServe(redisListenAddr)
}

const (
	redisListenAddr = ":8099"
)
