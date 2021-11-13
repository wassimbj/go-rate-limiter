package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisOpts struct {
	Addr     string // <host>:<port>
	Password string
}

type RLOpts struct {
	Attempts    int
	Prefix      string
	Duration    time.Duration // in seconds
	Id          string        // user identifier
	RedisConfig RedisOpts
}

func GetRedis(opts RedisOpts) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
	})

	if rdb.Ping(context.Background()).Err() != nil {
		return nil, rdb.Ping(context.Background()).Err()
	}

	return rdb, nil
	// err := rdb.Set(ctx, "key", "value", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }
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
		log.Fatalf("Redis Connection Error ! %s", err.Error())
		// nohting
		return RLResult{
			AttemptsLeft: 0,
			Used:         0,
			TimeLeft:     0,
			Block:        false,
		}, err
	}

	// construct the key "gorl:{prefix}:{id}", prefix ∈ (login, signup...) && id is a unique one can be an IP, userId...
	key := "gorl:" + opts.Prefix + opts.Id

	lock := NewLock(rds)
	lockId := lock.Acquire(ctx, opts.Prefix, time.Second*5)
	log.Println("LOCK ACQUIRED: ", lockId)

	data := rds.Get(ctx, key)

	attemptsLeft, _ := data.Int()
	timeLeft := rds.TTL(ctx, key).Val()

	fmt.Println("ATTEMPTS LEFT: ", attemptsLeft)

	// no data found, either the attempts expired or its the first time this user is making the request.
	if attemptsLeft == 0 && timeLeft < 0 {
		setResult := rds.Set(ctx, key, opts.Attempts-1, opts.Duration)
		log.Println("LOCK RELEASED (INIT): ", lock.Release(ctx, opts.Prefix, lockId))

		if setResult.Err() != nil {
			log.Fatalf("INIT ERROR %s", setResult.Err().Error())
			return RLResult{}, setResult.Err()
		}
		attemptsLeft = opts.Attempts - 1
		return RLResult{
			AttemptsLeft: opts.Attempts - 1,
			Used:         1,
			TimeLeft:     opts.Duration,
			Block:        false,
		}, nil
		// allow
	} else {
		if attemptsLeft <= 0 {
			// block user
			log.Println("LOCK RELEASED (BLOCK): ", lock.Release(ctx, opts.Prefix, lockId))
			return RLResult{
				AttemptsLeft: 0,
				Used:         opts.Attempts,
				TimeLeft:     time.Duration(timeLeft.Seconds()),
				Block:        true,
			}, nil
		} else {
			// update the attempts left
			rds.Decr(ctx, key)
			attemptsLeft = attemptsLeft - 1

			log.Println("LOCK RELEASED (UPDATE): ", lock.Release(ctx, opts.Prefix, lockId))

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

func main() {
	for j := 0; j < 50; j++ {
		go func() {
			result, _ := RateLimiter(context.Background(), RLOpts{
				Attempts: 5,
				Prefix:   "login",
				Duration: time.Second * 20,
				Id:       "A300",
				RedisConfig: RedisOpts{
					Addr: "localhost:6379",
					// Password: "",
				},
			})
			if result.Block {
				fmt.Println(j, " - BLOCKED !! TRY LATER")
				fmt.Println(result)
			} else {
				fmt.Println(j, " - Alright cool...")
				fmt.Println(result)
			}
		}()
	}

	time.Sleep(time.Hour)

}
