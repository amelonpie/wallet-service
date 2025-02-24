package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

// func TestMain(m *testing.M) {
// 	goleak.VerifyTestMain(m)
// }

//nolint:ireturn // stick to interface
func setupMockDB() (Repository, sqlmock.Sqlmock) {
	db, mockSQL, err := sqlmock.New() // mock db
	if err != nil {
		panic(err)
	}

	repo := newWalletRepository(db)

	return repo, mockSQL
}

/* Normal case */
func TestDeposit(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 100.00
	newBalance := 200.00

	// Assert
	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(newBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(userID, nil, amount, "deposit").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	// Act
	updatedBalance, err := repo.Deposit(context.Background(), userID, amount)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedBalance != newBalance {
		t.Fatalf("expected balance to be %v, got %v", newBalance, updatedBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestWithdraw(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 50.00
	newBalance := 150.00

	// Assert
	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(newBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(userID, nil, amount, "withdraw").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	// Act
	updatedBalance, err := repo.Withdraw(context.Background(), userID, amount)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updatedBalance != newBalance {
		t.Fatalf("expected balance to be %v, got %v", newBalance, updatedBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestTransfer(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	fromUserID := 1
	toUserID := 2
	amount := 30.00
	fromNewBalance := 20.00
	toNewBalance := 50.00

	mockSQL.ExpectBegin()

	// Assert: withdraw from user 1
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, fromUserID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(fromNewBalance))

	// Assert: deposit to user 2
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, toUserID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(toNewBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(fromUserID, toUserID, amount, "transfer").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockSQL.ExpectCommit()

	// Act
	fromBalance, toBalance, err := repo.Transfer(context.Background(), fromUserID, toUserID, amount)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fromBalance != fromNewBalance {
		t.Fatalf("expected from user balance to be %v, got %v", fromNewBalance, fromBalance)
	}

	if toBalance != toNewBalance {
		t.Fatalf("expected to user balance to be %v, got %v", toNewBalance, toBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetBalance(t *testing.T) {
	repo, mockSQL := setupMockDB()

	userID := 1
	expectedBalance := 100.00

	// Assert
	mockSQL.ExpectQuery(`SELECT balance FROM wallets WHERE user_id=\$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(expectedBalance))

	// Act
	balance, err := repo.GetBalance(context.Background(), userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance != expectedBalance {
		t.Fatalf("expected balance to be %v, got %v", expectedBalance, balance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetTransactionHistory(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	expectedTransactions := []Transaction{
		{
			TransactionID:   1,
			FromUserID:      1,
			ToUserID:        2,
			Amount:          100.00,
			TransactionType: "deposit",
			Timestamp:       time.Now().String(),
		},
		{
			TransactionID:   2,
			FromUserID:      1,
			ToUserID:        2,
			Amount:          50.00,
			TransactionType: "withdrawal",
			Timestamp:       time.Now().String(),
		},
	}

	// Assert
	mockSQL.ExpectQuery(`SELECT transaction_id, from_user_id, to_user_id, amount, transaction_type, timestamp FROM transactions WHERE from_user_id = \$1 OR to_user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "from_user_id", "to_user_id", "amount", "transaction_type", "timestamp"}).
			AddRow(1, 1, sql.NullInt64{Int64: 2, Valid: true}, 100.00, "deposit", time.Now().String()).
			AddRow(2, 1, sql.NullInt64{Int64: 2, Valid: true}, 50.00, "withdrawal", time.Now().String()))

	// Act
	transactions, err := repo.GetTransactionHistory(context.Background(), userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(transactions) != len(expectedTransactions) {
		t.Fatalf("expected %d transactions, got %d", len(expectedTransactions), len(transactions))
	}

	if transactions[0].Amount != expectedTransactions[0].Amount {
		t.Fatalf("expected amount to be %v, got %v", expectedTransactions[0].Amount, transactions[0].Amount)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

/* Abnormal case */
// Deposit
func TestDeposit_Error(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 100.00

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	_, err := repo.Deposit(context.Background(), userID, amount)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected error %v, got %v", sql.ErrConnDone, err)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestDeposit_ErrorInCommit(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 100.00
	newBalance := 200.00

	mockSQL.ExpectBegin()

	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(newBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(userID, nil, amount, "deposit").
		WillReturnResult(sqlmock.NewResult(1, 1))

	commitErr := errors.New("commit error")
	mockSQL.ExpectCommit().WillReturnError(commitErr)

	// Act
	updatedBalance, err := repo.Deposit(context.Background(), userID, amount)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := fmt.Errorf("failed to commit transaction for user %d: commit error", userID)
	if err.Error() != expectedErr.Error() {
		t.Fatalf("expected: %v, got: %v", expectedErr, err)
	}

	if updatedBalance != 0 {
		t.Fatalf("expected updated balance to be 0, got: %v", updatedBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestDeposit_ErrorInLogTransaction(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 100.00
	newBalance := 200.00

	// Assert
	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(newBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(userID, nil, amount, "deposit").
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	updatedBalance, err := repo.Deposit(context.Background(), userID, amount)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if updatedBalance != 0 {
		t.Fatalf("expected balance to be %v, got %v", newBalance, updatedBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestWithdraw_Error(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1
	amount := 50.00

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	_, err := repo.Withdraw(context.Background(), userID, amount)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected error %v, got %v", sql.ErrConnDone, err)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestTransfer_Error(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	fromUserID := 1
	toUserID := 2
	amount := 30.00

	// Mock connection error
	mockSQL.ExpectBegin()
	// withdraw from user 1
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, fromUserID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	_, _, err := repo.Transfer(context.Background(), fromUserID, toUserID, amount)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected error %v, got %v", sql.ErrConnDone, err)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetBalance_Error(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1

	mockSQL.ExpectQuery(`SELECT balance FROM wallets WHERE user_id=\$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)
	// Act
	_, err := repo.GetBalance(context.Background(), userID)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected error %v, got %v", sql.ErrConnDone, err)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetTransactionHistory_Error(t *testing.T) {
	// Arrange
	repo, mockSQL := setupMockDB()

	userID := 1

	mockSQL.ExpectQuery(`SELECT transaction_id, from_user_id, to_user_id, amount, transaction_type, timestamp FROM transactions WHERE from_user_id = \$1 OR to_user_id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	// Act
	_, err := repo.GetTransactionHistory(context.Background(), userID)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected error %v, got %v", sql.ErrConnDone, err)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestGetTransactionHistory_ScanError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &walletRepository{db: db}

	userID := 1
	queryRegex := `(?s)SELECT transaction_id, from_user_id, to_user_id, amount, transaction_type, timestamp FROM transactions WHERE from_user_id = \$1 OR to_user_id = \$1 ORDER BY timestamp DESC`
	columns := []string{"transaction_id", "from_user_id", "to_user_id", "amount", "transaction_type", "timestamp"}
	// Trigger error by rows.Scan, transaction_id should be a int, but now it is a string and cannot convert to int
	rows := sqlmock.NewRows(columns).
		AddRow("invalid", 1, 2, 100.0, "deposit", "2020-01-01T00:00:00Z")
	mock.ExpectQuery(queryRegex).WithArgs(userID).WillReturnRows(rows)

	// Act
	txs, err := repo.GetTransactionHistory(context.Background(), userID)

	// Assert
	if err == nil {
		t.Fatalf("expected error during row.Scan, got nil")
	}

	if txs != nil {
		t.Fatalf("expected no transactions, got: %v", txs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}
