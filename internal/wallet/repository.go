package wallet

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amelonpie/wallet-service/internal/database"
	"github.com/amelonpie/wallet-service/pkg/log"
)

//nolint:ireturn // stick to interface
func InitRepository() (Repository, error) {
	dbConfig := database.NewDatabaseConfig()

	postgre, err := dbConfig.ConnectPostgre()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		postgre.Close()

		os.Exit(0)
	}()

	return newWalletRepository(postgre), nil
}

// I think should return Repository as interface, not to return the concret type *walletRepository
// but the linter requires me to do so. Should I just ignore it?

//nolint:ireturn // stick to interface
func newWalletRepository(db *sql.DB) Repository {
	return &walletRepository{
		db:     db,
		logger: log.NewLogger("wallet").WithField("module", "endpoints"),
	}
}

func (r *walletRepository) handleTransaction(ctx context.Context, userID int, amount float64, query string, transactionType string) (float64, error) {
	var err error
	errptr := &err

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		errptr = &err
		return 0, fmt.Errorf("failed to begin transaction for user %d: %w", userID, err)
	}

	defer func() {
		if *errptr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.WithField("err", rollbackErr).Error("failed to roll back")
			}
		}
	}()

	var newBalance float64
	// Update the wallet balance based on the provided query
	err = tx.QueryRowContext(ctx, query, amount, userID).Scan(&newBalance)
	if err != nil {
		errptr = &err
		return 0, fmt.Errorf("failed to query database for user %d: %w", userID, err)
	}

	// Log the transaction within the same transaction
	err = r.LogTransaction(ctx, tx, &userID, nil, amount, transactionType)
	if err != nil {
		errptr = &err
		return 0, fmt.Errorf("failed to log transaction for user %d: %w", userID, err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		errptr = &err
		return 0, fmt.Errorf("failed to commit transaction for user %d: %w", userID, err)
	}

	return newBalance, nil
}

// Deposit adds the given amount to the user's wallet
func (r *walletRepository) Deposit(ctx context.Context, userID int, amount float64) (float64, error) {
	query := `UPDATE wallets SET balance = balance + $1 WHERE user_id = $2 RETURNING balance`
	return r.handleTransaction(ctx, userID, amount, query, "deposit")
}

// Withdraw subtracts the given amount from the user's wallet
func (r *walletRepository) Withdraw(ctx context.Context, userID int, amount float64) (float64, error) {
	query := `UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 RETURNING balance`
	return r.handleTransaction(ctx, userID, amount, query, "withdraw")
}

// Transfer moves `amount` from `fromUserID` to `toUserID`
func (r *walletRepository) Transfer(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	errptr := &err

	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction for from user %d and to user %d: %w", fromUserID, toUserID, err)
	}

	defer func() {
		if *errptr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.WithField("err", rollbackErr).Error("failed to roll back")
			}
		}
	}()

	// Subtract amount from `fromUserID`
	var fromBalance float64

	queryFrom := `UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 RETURNING balance`
	err = tx.QueryRowContext(ctx, queryFrom, amount, fromUserID).Scan(&fromBalance)

	if err != nil {
		errptr = &err
		return 0, 0, fmt.Errorf("failed to query database for user %d: %w", fromUserID, err)
	}

	// Add amount to `toUserID`
	var toBalance float64

	queryTo := `UPDATE wallets SET balance = balance + $1 WHERE user_id = $2 RETURNING balance`
	err = tx.QueryRowContext(ctx, queryTo, amount, toUserID).Scan(&toBalance)

	if err != nil {
		errptr = &err
		return 0, 0, fmt.Errorf("failed to query database for user %d: %w", toUserID, err)
	}

	// Log the transaction within the same transaction
	err = r.LogTransaction(ctx, tx, &fromUserID, &toUserID, amount, "transfer")
	if err != nil {
		errptr = &err
		return 0, 0, fmt.Errorf("failed to log transaction for from user %d and to %d: %w", fromUserID, toUserID, err)
	}

	if err = tx.Commit(); err != nil {
		errptr = &err
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return fromBalance, toBalance, nil
}

// GetBalance returns the current balance of the specified user
func (r *walletRepository) GetBalance(ctx context.Context, userID int) (float64, error) {
	var balance float64

	query := `SELECT balance FROM wallets WHERE user_id=$1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)

	if err != nil {
		return 0, fmt.Errorf("failed to query database for user %d: %w", userID, err)
	}

	return balance, nil
}

// LogTransaction inserts a new record into the transactions table
func (r *walletRepository) LogTransaction(
	ctx context.Context,
	tx *sql.Tx,
	fromUserID, toUserID *int,
	amount float64,
	transactionType string,
) error {
	query := `INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type)
              VALUES ($1, $2, $3, $4)`

	_, err := tx.ExecContext(ctx, query, fromUserID, toUserID, amount, transactionType)
	if err != nil {
		return fmt.Errorf("failed to insert database: %w", err)
	}

	return nil
}

// GetTransactionHistory retrieves transactions for a particular user
// Linter complaints about append without preallocation, but the exact number of rows
// need query twice. Assume large database so just ignore
// var count int
// err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(&count)
//
//	if err != nil {
//		return nil, err
//	}
func (r *walletRepository) GetTransactionHistory(ctx context.Context, userID int) ([]Transaction, error) {
	query := `
    SELECT transaction_id, from_user_id, to_user_id, amount, transaction_type, timestamp
    FROM transactions
    WHERE from_user_id = $1 OR to_user_id = $1
    ORDER BY timestamp DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query database for user %d: %w", userID, err)
	}
	defer rows.Close()

	//nolint:prealloc // see function start
	var txs []Transaction

	for rows.Next() {
		var t Transaction
		err = rows.Scan(
			&t.TransactionID,
			&t.FromUserID,
			&t.ToUserID,
			&t.Amount,
			&t.TransactionType,
			&t.Timestamp,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan certain transaction for user %d: %w", userID, err)
		}

		txs = append(txs, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error arises during rows intertation for user %d: %w", userID, err)
	}

	return txs, nil
}
