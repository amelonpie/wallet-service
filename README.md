# Wallet Service

## Install and configurate dependencies
### Redis
[quick start](https://redis.io/learn/howtos/quick-start)
```sh
docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
# connection string is redis://default:redispw@localhost:49153
# redis-cli
docker exec -it redis-nOxO redis-cli -a redispw
 127.0.0.1:6379> set mykey myvalue
OK
127.0.0.1:6379> get mykey
"myvalue" 
# delete docker container and its data volume
docker rm -v --force redis-nOxO
```
Now you can enter redis-cli commands such as `incr mycounter` to create / increment an integer counter, or `set mykey myvalue` to store the string `myvalue` against the key `mykey`.
### PostgreSQL
[tutorial](https://www.postgresql.org/docs/current/tutorial.html)
