# Example using go-rate-limiter

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/tomasen/realip"
	"github.com/go-redis/redis/v8"
)

func Login(res http.ResponseWriter, req *http.Request) {

	// get user ip
	USER_IP := realip.FromRequest(req)

	// call the rate limiter
	rl, _ := RateLimiter(context.Background(), RLOpts{
		Attempts: 10,
		Prefix:   "login",
		Duration: time.Hour * 5,
		Id:       USER_IP, // the id of the user who's making this request
		RedisConfig: redis.Options{
			Addr: "localhost:6379",
		},
	})

	// check if the user should get blocked
	if rl.Block {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusTooManyRequests)
		timeLeft := float32(rl.TimeLeft) / float32(time.Millisecond)
		res.Write([]byte(fmt.Sprintf("You reached the limit, come back after %vms", timeLeft)))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	attemptsLeft := strconv.Itoa(rl.AttemptsLeft)
	res.Write([]byte("Attempts Left: " + attemptsLeft))
	// log.Println("Login user: ", clientIP)
}

func main() {

	http.HandleFunc("/login", Login)

	log.Println("SEVER IS RUNNIG")

	http.ListenAndServe(":1234", nil)
}
```
