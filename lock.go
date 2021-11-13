package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

type Lock struct {
	redis *redis.Client
}

func NewLock(rds *redis.Client) *Lock {
	return &Lock{
		redis: rds,
	}
}

func RandToken(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letters = []rune("abcdefghijklmnopqrstuvwxyz-_:!$ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (l *Lock) Acquire(ctx context.Context, lockname string, expiration time.Duration) string {
	lockId := RandToken(7)
	key := "lock:" + lockname
	tick := time.NewTicker(time.Nanosecond * 10)

	// try acquiring the lock until it times-out
	for {
		select {
		case <-tick.C: // each 1ms try to acquire the lock
			setnxCmd := l.redis.SetNX(ctx, key, lockId, expiration)
			if ok, _ := setnxCmd.Result(); ok {
				return lockId
			}
			continue
		case <-time.After(time.Second * 10): // time out
			fmt.Println("TIMED OUT ! GET OUT")
			return ""
		}
	}
}

func (l *Lock) Release(ctx context.Context, lockname, lockId string) bool {

	key := "lock:" + lockname

	txf := func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			getCmd := pipe.Get(ctx, key)

			// delete the lock if its found
			if getCmd.Val() == lockId {
				pipe.Del(ctx, key)
			}
			return nil
		})
		return err
	}

	for {
		err := l.redis.Watch(ctx, txf, key)
		if err == nil {
			return true
		} else if err == redis.TxFailedErr {
			// we lost the lock, retry !
			continue
		} else {
			return false
		}
	}

}
