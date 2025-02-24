# Review

## Structure
```
.
├── Dockerfile
├── README.md
├── Review.md
├── cmd
│   └── wallet_service
│       └── main.go
├── config
│   └── config.yaml
├── configs
│   ├── config.yaml
│   └── init.sql
├── docker-compose.yaml
├── go.mod
├── go.sum
├── installGolangci-lint.sh
├── internal
│   ├── database  // methods to connect to PostgreSQL and Redis
│   │   ├── config.go
│   │   ├── postgresql.go
│   │   └── redis.go
│   ├── endpoint  // request router and handler
│   │   ├── endpoint.go
│   │   ├── model.go
│   │   ├── router.go
│   │   ├── transaction.go
│   │   ├── transaction_test.go
│   │   ├── view.go
│   │   └── view_test.go
│   └── wallet  // core business logic
│       ├── error.go
│       ├── model.go
│       ├── repository.go  // layer directly connecting to PostgreSQL
│       ├── repository_test.go
│       ├── service.go  // layer querying Redis and fallback to PostgreSQL
│       └── service_test.go
└── pkg  // previous tools
    ├── config
    │   └── config.go
    └── log
        └── log.go
```

## Unit test
#### Coverage
The current coverage is below. From my understanding, only public methods need unit test. In this way, some branches of error handling are be hard to reach.

It may be the places where I should improve myself, to learn how to design codes that make unit test reachable as much as possible. Maybe if there are more efforts that will come a way.

Any suggestions? Will greatly appreciate your review.
```sh
ok      github.com/amelonpie/wallet-service/internal/endpoint   1.247s  coverage: 62.0% of statements
ok      github.com/amelonpie/wallet-service/internal/wallet     1.337s  coverage: 73.6% of statements
```

## Time spent
Started from 20 Feb understanding the requirements, search necessary tutorials, design layers,
21 - 24 working on defining types, implementations
24 - 25: fix linter, refactor, work on docker

## No bonus