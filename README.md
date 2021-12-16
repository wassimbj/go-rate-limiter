# Go rate limiter

## Features

- Simple
- Fast
- Distributed
- Uses Token bucket algorithm
- Redis Based
- Flexible
- Can handle concurrent requests

## Install

```bash
go get github.com/wassimbj/gorl
```

## Usage

```go
import "github.com/wassimbj/gorl"


type Result struct {
	AttemptsLeft int
	Used         int           // used attempts
	TimeLeft     int           // in ms, time left until the attempts gets renewed
	Block        bool          // should the user get blocked
}

// in this example, the user has the right to make only 50 requests in 5 hours.
result, err := gorl.RateLimiter(context.Background(), gorl.RLOpts{
   Attempts: 50, // requests limit
   Prefix:   "login",
   Duration: time.Hour * 5, // limit duration
   Id:       "USER_IP",
   BlockDuration: time.Hour,
   RedisClient: redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}),
})

```

**Note:** the bucket(redis) is refilled with limit each time the duration ends, im not saving the last refill timestamp here maybe in near future.

## Full example

Go to **[example.md](https://github.com/wassimbj/go-rate-limiter/blob/master/example.md)**

## TODO

:white_check_mark: Add block duration
