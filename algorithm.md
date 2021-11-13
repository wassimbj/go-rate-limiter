
### Algorithm

```go
Token bucket algorithm:

attempts = 10
key = "rate:login:{USER_IDENTIFIER}" | USER_IDENTIFIER ∈ (ip, id, ...)
duration = "1 hour"
bucket = [] // this is redis


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
         bucket.decrement(key)

```
