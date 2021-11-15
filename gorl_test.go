package gorl

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func RdsClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "0.0.0.0:6379",
	})
}

func TestRateLimiter(t *testing.T) {
	var wg sync.WaitGroup

	for j := 0; j < 20; j++ {
		wg.Add(1)
		go func(c int) {
			res, err := RateLimiter(context.Background(), RLOpts{
				Attempts:    10,
				Prefix:      "login",
				Duration:    time.Second * 30,
				Id:          "A300",
				RedisClient: RdsClient(),
			})

			if err != nil {
				t.Log(err)
				t.Fail()
			}
			fmt.Println(fmt.Sprintf("%d - %dms", res.AttemptsLeft, res.TimeLeft))

			defer wg.Done()
		}(j)
	}

	wg.Wait()
}

func TestRandToken(t *testing.T) {
	s := RandToken(10)

	if len(s) != 10 {
		t.Fail()
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	var wg sync.WaitGroup

	for j := 0; j < b.N; j++ {
		wg.Add(1)
		go func(c int) {
			RateLimiter(context.Background(), RLOpts{
				Attempts:    100,
				Prefix:      "login",
				Duration:    time.Hour * 5,
				Id:          "A300",
				RedisClient: RdsClient(),
			})

			defer wg.Done()
		}(j)
	}

	wg.Wait()
}
