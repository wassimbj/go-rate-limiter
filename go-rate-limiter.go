package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RLOpts struct {
	Attempts    int
	Prefix      string
	Duration    time.Duration // in seconds
	Id          string        // user identifier
	RedisConfig redis.Options
}

func GetRedis(opts redis.Options) (*redis.Client, error) {

	rdb := redis.NewClient(&opts)

	if rdb.Ping(context.Background()).Err() != nil {
		return nil, rdb.Ping(context.Background()).Err()
	}

	return rdb, nil
}

type RLResult struct {
	AttemptsLeft int
	Used         int           // used attempts
	TimeLeft     time.Duration // time left until the bucket gets refilled
	Block        bool          // should the user get blocked
}

func RateLimiter(ctx context.Context, opts RLOpts) (RLResult, error) {
	rds, err := GetRedis(opts.RedisConfig)

	if err != nil {
		// nohting
		return RLResult{
			AttemptsLeft: 0,
			Used:         0,
			TimeLeft:     0,
			Block:        false,
		}, err
	}

	// construct the key "gorl:{prefix}:{id}", prefix ∈ (login, signup...) && id is a unique one can be an IP, userId...
	key := fmt.Sprintf("gorl:%s:%s", opts.Prefix, opts.Id)

	lock := NewLock(rds)
	lockId := lock.Acquire(ctx, opts.Prefix, time.Second*2)
	defer func() {
		lock.Release(ctx, opts.Prefix, lockId)
	}()

	data := rds.Get(ctx, key)

	attemptsLeft, _ := data.Int()
	timeLeft := rds.TTL(ctx, key).Val()

	// no data found, either the attempts expired or its the first time this user is making the request.
	if attemptsLeft <= 0 && timeLeft < 0 {
		setResult := rds.Set(ctx, key, opts.Attempts-1, opts.Duration)

		if setResult.Err() != nil {
			// log.Fatalf("INIT ERROR %s", setResult.Err().Error())
			return RLResult{}, setResult.Err()
		}
		attemptsLeft = opts.Attempts - 1
		return RLResult{
			AttemptsLeft: attemptsLeft,
			Used:         1,
			TimeLeft:     opts.Duration,
			Block:        false,
		}, nil
		// allow
	} else {
		if attemptsLeft <= 0 {
			// block user
			return RLResult{
				AttemptsLeft: 0,
				Used:         opts.Attempts,
				TimeLeft:     time.Duration(timeLeft.Seconds()),
				Block:        true,
			}, nil
		} else {
			// update the attempts left
			decrCmd := rds.Decr(ctx, key)

			attemptsLeft = int(decrCmd.Val())

			// allow the user
			return RLResult{
				AttemptsLeft: attemptsLeft,
				Used:         opts.Attempts - attemptsLeft,
				TimeLeft:     time.Duration(timeLeft.Seconds()),
				Block:        false,
			}, nil
		}
	}

}
