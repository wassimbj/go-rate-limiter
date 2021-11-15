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
go get -d github.com/wassimbj/go-rate-limiter
```

## Usage

```go
// in this example, the user has the right to make only 50 requests in 5 hours. 
result, err := RateLimiter(context.Background(), RLOpts{
   Attempts: 50, // requests limit
   Prefix:   "login",
   Duration: time.Hour * 5, // limit duration
   Id:       "USER_IP",
   RedisConfig: redis.Options{
      Addr: "localhost:6379",
   },
})

/*
result = type RLResult struct {
	AttemptsLeft int
	Used         int           // used attempts
	TimeLeft     int           // in ms, time left until the attempts gets renewed
	Block        bool          // should the user get blocked
}
*/
```

**Note:** the bucket(redis) is refilled with limit each time the duration ends, im not saving the last refill timestamp here maybe in near future.


## Full example

Go to **[example.md](https://github.com/wassimbj/go-rate-limiter/blob/master/example.md)**
