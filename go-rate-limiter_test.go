package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRateLimiter(t *testing.T) {
	var wg sync.WaitGroup

	for j := 0; j < 20; j++ {
		wg.Add(1)
		go func(c int) {
			_, err := RateLimiter(context.Background(), RLOpts{
				Attempts: 50,
				Prefix:   "login",
				Duration: time.Hour * 5,
				Id:       "A300",
				RedisConfig: redis.Options{
					Addr: "localhost:6379",
				},
			})

			if err != nil {
				t.Log(err)
				t.Fail()
			}
			// fmt.Println(res.AttemptsLeft, "-", res.TimeLeft)

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

func TestLock(t *testing.T) {
	rds, _ := getRedis(redis.Options{
		Addr: "localhost:6379",
	})

	lock := NewLock(rds)
	lockId := lock.Acquire(context.Background(), "test", time.Second*10)
	if lockId == "" {
		t.Fail()
	}

	lock.Release(context.Background(), "test", lockId)

}

func BenchmarkRateLimiter(b *testing.B) {
	var wg sync.WaitGroup

	for j := 0; j < b.N; j++ {
		wg.Add(1)
		go func(c int) {
			RateLimiter(context.Background(), RLOpts{
				Attempts: 100,
				Prefix:   "login",
				Duration: time.Hour * 5,
				Id:       "A300",
				RedisConfig: redis.Options{
					Addr: "localhost:6379",
					// Password: "",
				},
			})

			defer wg.Done()
		}(j)
	}

	wg.Wait()
}
