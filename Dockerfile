FROM golang:1.21.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN  go env -w GOPROXY=https://goproxy.cn,direct && go mod tidy
COPY . .
RUN go build ./cmd/wallet_service

FROM alpine:3.18
WORKDIR /root/
COPY --from=builder /app/wallet_service .
EXPOSE 3000

# Command to run the executable
CMD ["/root/wallet_service"]
