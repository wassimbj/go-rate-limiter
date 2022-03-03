package gorl

import (
	"context"
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

	var letters = []rune("abcdefghijklmnopqrstu1234567890vwxyz-_:!$ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (l *Lock) Acquire(ctx context.Context, lockname string, expiration time.Duration) string {
	lockId := RandToken(10)
	key := "lock:" + lockname
	// tick := time.NewTicker(time.Nanosecond * 10)
	timer := time.NewTimer(time.Second * 10)
	// try acquiring the lock until it times-out or the key expire
	for {
		select {
		case <-timer.C: // time out
			timer.Stop()
			return ""
		default: // try to acquire the lock
			setNxCmd := l.redis.SetNX(ctx, key, lockId, expiration)
			ok, _ := setNxCmd.Result()
			if ok {
				return lockId
			}
			continue
		}
	}
}

func (l *Lock) Release(ctx context.Context, lockname, lockId string) bool {

	timer := time.NewTimer(time.Second * 50)
	key := "lock:" + lockname
	txf := func(tx *redis.Tx) error {
		getCmd := tx.Get(ctx, key)
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			// delete the lock if its found
			if getCmd.Val() == lockId {
				pipe.Del(ctx, key)
			}
			return nil
		})
		return err
	}

	for {
		select {
		case <-timer.C: // time out
			timer.Stop()
			return false
		default:
			err := l.redis.Watch(ctx, txf, key)
			if err == nil {
				return true
			} else if err == redis.TxFailedErr {
				// something wrong, we either lost the key or an unexpected thing happened, just try again
				continue
			} else {
				return false
			}

		}
	}

}
