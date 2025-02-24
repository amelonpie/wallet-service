package wallet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amelonpie/wallet-service/internal/database"
	"github.com/redis/go-redis/v9"
)

//nolint:ireturn // stick to interface
func InitService(repo Repository) (Service, error) {
	dbConfig := database.NewDatabaseConfig()

	cache, err := dbConfig.ConnectRedis()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return newWalletService(repo, cache), nil
}

//nolint:ireturn // stick to interface
func newWalletService(repo Repository, cache *redis.Client) Service {
	return &walletService{
		repo:  repo,
		cache: cache,
	}
}

func (s *walletService) Deposit(ctx context.Context, userID int, amount float64) (float64, error) {
	newBalance, err := s.repo.Deposit(ctx, userID, amount)
	if err != nil {
		return 0, fmt.Errorf("failed to deposit to database for user %d: %w", userID, err)
	}

	// Update Redis cache
	key := fmt.Sprintf("wallet_balance:%d", userID)
	err = s.cache.Set(ctx, key, newBalance, 0).Err()

	if err != nil {
		return 0, fmt.Errorf("failed to update redis for user %d: %w", userID, err)
	}

	return newBalance, nil
}

func (s *walletService) Withdraw(ctx context.Context, userID int, amount float64) (float64, error) {
	// Check balance
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	if balance < amount {
		return 0, errorFactory().InsufficientFunds
	}

	newBalance, err := s.repo.Withdraw(ctx, userID, amount)
	if err != nil {
		return 0, fmt.Errorf("failed to update database for user %d: %w", userID, err)
	}

	// Update Redis cache
	key := fmt.Sprintf("wallet_balance:%d", userID)
	err = s.cache.Set(ctx, key, newBalance, 0).Err()

	if err != nil {
		return 0, fmt.Errorf("failed to update redis for user %d: %w", userID, err)
	}

	return newBalance, nil
}

func (s *walletService) Transfer(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error) {
	// Check fromUserID balance
	fromBalance, err := s.GetBalance(ctx, fromUserID)
	if err != nil {
		return 0, 0, err
	}

	if fromBalance < amount {
		return 0, 0, errorFactory().InsufficientFunds
	}

	_, err = s.GetBalance(ctx, toUserID)
	if err != nil {
		return 0, 0, errorFactory().RecipientNotFound
	}

	newFromBalance, newToBalance, err := s.repo.Transfer(ctx, fromUserID, toUserID, amount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update database: %w", err)
	}

	// Update Redis caches
	fromKey := fmt.Sprintf("wallet_balance:%d", fromUserID)
	toKey := fmt.Sprintf("wallet_balance:%d", toUserID)

	// Update fromUser
	err = s.cache.Set(ctx, fromKey, newFromBalance, 0).Err()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update redis: %w", err)
	}

	// Update toUser
	err = s.cache.Set(ctx, toKey, newToBalance, 0).Err()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update redis: %w", err)
	}

	return newFromBalance, newToBalance, nil
}

func (s *walletService) GetBalance(ctx context.Context, userID int) (float64, error) {
	// Check Redis first
	key := fmt.Sprintf("wallet_balance:%d", userID)
	cachedBalance, err := s.cache.Get(ctx, key).Result()

	if err == nil {
		// If found in cache, parse to float
		var balance float64
		balance, err = strconv.ParseFloat(cachedBalance, 64)

		if err == nil {
			return balance, nil
		}
	}

	// Fallback to Postgres
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance to get balance for user %d: %w", userID, err)
	}

	// Update cache for next time
	err = s.cache.Set(ctx, key, balance, 0).Err()
	if err != nil {
		return 0, fmt.Errorf("failed to update redis for user %d: %w", userID, err)
	}

	return balance, nil
}

func (s *walletService) GetTransactionHistory(ctx context.Context, userID int) ([]Transaction, error) {
	// For now, no cache. Read directly from DB:
	txs, err := s.repo.GetTransactionHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history for user %d: %w", userID, err)
	}

	return txs, nil
}
