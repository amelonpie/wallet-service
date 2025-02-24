# Wallet Service

## Build Dockerfile and docker compose
In case docker environment is not available, refer to section [Install and configurate dependencies](#install-and-configurate-dependencies).
```sh
docker build -t wallet_service .
docker compose up
```

## Test
### Unit test
Only test the internal codes. No need to test package database as it needs to connect to real post gre or redis.
** improvement: how to mock?
```sh
go test ./internal/... -race -cover
```
To test the goleak, comment out line 15-17 of `internal/wallet/repository_test.go`, and line 16-18 of `internal/endpoint/transaction_test.go`

### API test
TBD: Postman
#### curl
##### Deposit
```sh
curl --request POST \
  --url http://localhost:3000/wallet/1/deposit \
  --header 'Content-Type: application/json' \
  --data '{"user_id": 1, "amount": 50.0}'
# should receive:
# {"new_balance":150,"status":"success"}
# api should log
# {"amount":50,"file":"/mnt/e/wallet-service/internal/endpoint/transaction.go:171","func":"github.com/amelonpie/wallet-service/internal/endpoint.handleTransactionRequest","level":"info","module":"endpoints","msg":"successful deposit","newBalance":150,"time":"2025-02-25T02:46:19+08:00","user_id":1}
```

##### Withdraw
```sh
curl --request POST \
  --url http://localhost:3000/wallet/1/withdraw \
  --header 'Content-Type: application/json' \
  --data '{"user_id": 1, "amount": 50.0}'
# should receive:
# {"new_balance":100,"status":"success"}
# api should log
# {"amount":50,"file":"/mnt/e/wallet-service/internal/endpoint/transaction.go:171","func":"github.com/amelonpie/wallet-service/internal/endpoint.handleTransactionRequest","level":"info","module":"endpoints","msg":"successful withdraw","newBalance":100,"time":"2025-02-25T02:48:34+08:00","user_id":1}
```

##### Transfer
```sh
curl --request POST \
  --url http://localhost:3000/wallet/transfer \
  --header 'Content-Type: application/json' \
  --data '{"from_user_id": 1, "to_user_id": 2, "amount": 10.0}'
# should receive
# {"from_balance":90,"status":"success","to_balance":60}
# api should log
# {"amount":10,"file":"/mnt/e/wallet-service/internal/endpoint/transaction.go:226","from_user_id":1,"func":"github.com/amelonpie/wallet-service/internal/endpoint.transferHandler","level":"info","module":"endpoints","msg":"successful transfer","new_from_balance":90,"new_to_balance":60,"time":"2025-02-25T02:55:36+08:00","to_user_id":2}
```

##### Get balance
```sh
curl http://localhost:3000/wallet/1/balance
# should receive
# {"balance":90}
# api should log
# {"balance":90,"file":"/mnt/e/wallet-service/internal/endpoint/view.go:68","func":"github.com/amelonpie/wallet-service/internal/endpoint.balanceHandler","level":"info","module":"endpoints","msg":"successful get balance","time":"2025-02-25T03:01:58+08:00","user_id":1}
```

##### Get transaction history
```sh
# test user 2. user 1 has too long history
curl http://localhost:3000/wallet/2/transactions
# should receive
# [{"transaction_id":3,"from_user_id":1,"to_user_id":{"Int64":2,"Valid":true},"amount":10,"transaction_type":"transfer","timestamp":"2025-02-24T18:51:56.682079Z"}]
# api should log
# {"file":"/mnt/e/wallet-service/internal/endpoint/view.go:109","func":"github.com/amelonpie/wallet-service/internal/endpoint.transactionsHandler","level":"info","module":"endpoints","msg":"successful get transaction history","time":"2025-02-25T03:13:02+08:00","transaction":[{"transaction_id":3,"from_user_id":1,"to_user_id":{"Int64":2,"Valid":true},"amount":10,"transaction_type":"transfer","timestamp":"2025-02-24T18:51:56.682079Z"}],"user_id":2}
```

## CI
### lint
Only test the internal codes. No
```sh
# install
./installGolangci-lint.sh
golangci-lint run -c .golangci.yaml ./internal/wallet/*
```
There are some stupid suggestions (false positive) for the linter and more time is needed to investigate how to satisfy it.
For example, `internal/endpoint/view.go:1:1: File is not properly formatted (gci)`, but when I use `gci` to fix this file, it simply adds new line between every line.

### leak
Goleak will complain the connection to PostgreSQL does not close after test. In the implementation, the connection to database keep alive until program terminated then call `postgres.Close()` which will also cause a Goleak.

Better design should be considered on handling database connection close.
```sh
Goroutine 37 in state select, with database/sql.(*DB).connectionOpener on top of the stack:
database/sql.(*DB).connectionOpener(0xc000264270, {0xd66378, 0xc000140c80})
        /usr/local/go/src/database/sql/sql.go:1261 +0xeb
created by database/sql.OpenDB in goroutine 36
        /usr/local/go/src/database/sql/sql.go:841 +0x287
```

## Install and configurate dependencies
### Redis
```sh
docker pull redis/redis-stack:latest
# or docker pull docker.1ms.run/redis/redis-stack:latest for proxy
docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
```
For manual testing purpose, there is no pre-defined information in the redis.

### PostgreSQL
```sh
docker pull postgres
# or docker pull docker.1ms.run/postgres for proxy
docker run -d --name mypostgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=yourpassword -e POSTGRES_DB=users postgres -c 'ssl=off'
# use psql to insert example data
docker exec -it mypostgres psql -U postgres
# create table users
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
);
# insert demo user
INSERT INTO users (username, email)
VALUES ('john', 'john@example.com');
INSERT INTO users (username, email)
VALUES ('tom', 'tom@example.com');
# should see there are two users
SELECT * FROM users;
 user_id | username |        email         
---------+----------+----------------------
       1 | john_doe | john.doe@example.com
       2 | tom      | tom@example.com
(2 rows)

# now create table wallet
CREATE TABLE IF NOT EXISTS wallets (
    wallet_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    balance DECIMAL(15, 2) DEFAULT 0.00
);
# insert example
INSERT INTO wallets (user_id, balance)
VALUES
    (1, 100.00),
    (2, 50.00);
# should get 
 wallet_id | user_id | balance 
-----------+---------+---------
         1 |       1 |  100.00
         2 |       2 |   50.00

# create transaction table
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id SERIAL PRIMARY KEY,
    from_user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    to_user_id INT REFERENCES users(user_id),
    amount DECIMAL(15, 2),
    transaction_type VARCHAR(20), -- 'deposit', 'withdrawal', 'transfer'
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
Then we will use API test to generate real transactions.

## Build and run the program
```sh
go build -v ./cmd/wallet_service/
./wallet_service -c configs/config.yaml 
```