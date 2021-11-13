# Go rate limiter

### Algorithm

```go
Token bucket algorithm:

attempts = 10
key = "rate:login:{IP}"
duration = "1 hour"
bucket = [
   key_1 attempts_1 EXPIRE duration_1,
   key_2 attempts_2 EXPIRE duration_2,
   key_3 attempts_3 EXPIRE duration_3,
]


while(requests_are_coming_in)

   data = bucket.get(key)

   if data.attempts == 0 && data.timeLeft < 0 then
      //? the attempts expired, or its the first time the user making the request.
      bucket.init(key, attempts, EX duration)
   else
      //? we ran out of attempts
      if data.attempts == 0 then
         user.block()
      else:
         user.allow()
         bucket.update(key, attempts--, EX duration)

```
