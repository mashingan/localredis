# localredis

[![Go](https://github.com/mashingan/localredis/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/mashingan/localredis/actions/workflows/go.yml)

Local implementation of redis protocol. The storage in entirely put in the memory and will be lost
if the app/program is quitting.  
This implementation usually used for integration testing without having
to run actual redis server as this is easy to run and close as the test starts and ends.  

It's main usage why re-invented the wheel is for ease of mock testing without having fully-blown
redis server installation. Mostly the usage can be summarized as below example:

```go
package testing
import (
    "APP-ROOT/app"
    "testing"
    "os"

    "github.com/mashingan/localredis"
)
func checkAppStates(t *testing.T, app.App) {
    // implementation
}

func TestOurApp(t *testing.T) {
    redisAddr := ":8099"
    os.Setenv("REDIS_ADDR", redisAddr)
    go localredis.ListenAndServe(redisAddr)
    apprun := app.Start()
    checkAppStates(t, apprun)
    localredis.Close()
}
```

Above example taking assumption that the app would connect to redis from `REDIS_ADDR` environment variable value.  
So instead of connecting to actual redis server, the app connect to our in memory redis.

Currently, it's in alpha-state with only basic `set`, `get` and `ping` handler implemented.

# Install

Using go modules, simply importing the path [`github.com/mashingan/localredis`](github.com/mashingan/localredis)
will automatically put it on `go.mod` dependencies entry.

# Contributing

Adding supports for the new command by adding it to [`commandsMap`](commands.go#L42).

# License

MIT