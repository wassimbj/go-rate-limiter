package gorl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RLOpts struct {
	Attempts    int
	Prefix      string
	Duration    time.Duration
	Id          string // user identifier
	RedisClient *redis.Client
}

type RLResult struct {
	AttemptsLeft int
	Used         int   // used attempts
	TimeLeft     int64 // in ms, time left until the bucket gets refilled
	Block        bool  // should the user get blocked
}

func RateLimiter(ctx context.Context, opts RLOpts) (RLResult, error) {
	redisClient := opts.RedisClient

	if redisClient.Ping(ctx).Val() != "PONG" {
		connErr := errors.New("REDIS CONNECTION ERROR ! " + redisClient.Ping(ctx).Err().Error())
		return RLResult{}, connErr
	}

	// construct the key "gorl:{prefix}:{id}", prefix ∈ (login, signup...) && id is a unique one can be an IP, userId...
	key := fmt.Sprintf("gorl:%s:%s", opts.Prefix, opts.Id)
	lockName := fmt.Sprintf("%s:%s", opts.Prefix, opts.Id)

	lock := NewLock(redisClient)
	lockId := lock.Acquire(ctx, lockName, time.Second*5)
	defer lock.Release(ctx, lockName, lockId)

	data := redisClient.Get(ctx, key)

	attemptsLeft, _ := data.Int()
	timeLeft := redisClient.PTTL(ctx, key).Val()

	// no data found, either the attempts expired or its the first time this user is making the request.
	if attemptsLeft <= 0 && timeLeft < 0 {
		setResult := redisClient.Set(ctx, key, opts.Attempts-1, opts.Duration)

		if setResult.Err() != nil {
			// log.Fatalf("INIT ERROR %s", setResult.Err().Error())
			return RLResult{}, setResult.Err()
		}
		attemptsLeft = opts.Attempts - 1
		return RLResult{
			AttemptsLeft: attemptsLeft,
			Used:         1,
			TimeLeft:     opts.Duration.Milliseconds(),
			Block:        false,
		}, nil
		// allow
	} else {
		if attemptsLeft <= 0 {
			// block user
			return RLResult{
				AttemptsLeft: 0,
				Used:         opts.Attempts,
				TimeLeft:     timeLeft.Milliseconds(),
				Block:        true,
			}, nil
		} else {
			// update the attempts left
			decrCmd := redisClient.Decr(ctx, key)

			attemptsLeft = int(decrCmd.Val())

			// allow the user
			return RLResult{
				AttemptsLeft: attemptsLeft,
				Used:         opts.Attempts - attemptsLeft,
				TimeLeft:     timeLeft.Milliseconds(),
				Block:        false,
			}, nil
		}
	}

}
