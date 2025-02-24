package wallet

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Transaction struct is used to map records from transactions table
type Transaction struct {
	TransactionID   int     `json:"transaction_id"`
	FromUserID      int     `json:"from_user_id"`
	ToUserID        int     `json:"to_user_id,omitempty"`
	Amount          float64 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	Timestamp       string  `json:"timestamp"`
}

// Repository defines methods to interact with the wallet data.
type Repository interface {
	Deposit(ctx context.Context, userID int, amount float64) (float64, error)
	Withdraw(ctx context.Context, userID int, amount float64) (float64, error)
	Transfer(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error)
	GetBalance(ctx context.Context, userID int) (float64, error)
	LogTransaction(ctx context.Context, tx *sql.Tx, fromUserID, toUserID *int, amount float64, transactionType string) error
	GetTransactionHistory(ctx context.Context, userID int) ([]Transaction, error)
}

type Service interface {
	Deposit(ctx context.Context, userID int, amount float64) (float64, error)
	Withdraw(ctx context.Context, userID int, amount float64) (float64, error)
	Transfer(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error)
	GetBalance(ctx context.Context, userID int) (float64, error)
	GetTransactionHistory(ctx context.Context, userID int) ([]Transaction, error)
}

type walletService struct {
	repo  Repository
	cache *redis.Client
}

type walletRepository struct {
	db     *sql.DB
	logger *logrus.Entry
}
