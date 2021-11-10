# Go rate limiter

### Algorithm

```go
Token bucket algorithm:

attempts = 10
key = "rate:login:{IP}"
duration = "1 hour"
bucket = key attempts EX duration


while(requests_are_coming_in)

   data = bucket.get(key)

   if !data then
      //? the attempts expired, or its the first time.
      bucket.init(key, attempts - 1, EX duration)
   else
      //? we ran out of attempts
      if data.attempts == 0 then
         user.block()
      else:
         user.allow()
         bucket.update(key, attempts--, EX duration)

```
