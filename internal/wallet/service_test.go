package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
)

//nolint:ireturn // stick to interface
func setupMockRepo() (Service, sqlmock.Sqlmock, redismock.ClientMock) {
	db, mockSQL, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	repo := newWalletRepository(db)
	mockRedisClient, mockRedis := redismock.NewClientMock()

	service := newWalletService(repo, mockRedisClient)

	return service, mockSQL, mockRedis
}

func TestWalletService_Deposit(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

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
	mockRedis.ExpectSet(`wallet_balance:1`, newBalance, 0).SetVal("OK")

	// Act
	updatedBalance, err := service.Deposit(context.Background(), userID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

func TestWalletService_Deposit_Error(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

	userID := 1
	amount := 100.00

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	_, err := service.Deposit(context.Background(), userID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

func TestWalletService_Withdraw(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

	userID := 1
	amount := 50.00
	newBalance := 150.00
	original := amount + newBalance

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(newBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(userID, nil, amount, "withdraw").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()
	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", userID)).SetVal(fmt.Sprintf("%f", original))
	mockRedis.ExpectSet(fmt.Sprintf("wallet_balance:%d", userID), newBalance, 0).SetVal("OK")

	// Act
	updatedBalance, err := service.Withdraw(context.Background(), userID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

func TestWalletService_Withdraw_Error(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

	userID := 1
	amount := 50.00
	newBalance := 150.00
	original := amount + newBalance

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, userID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()
	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", userID)).SetVal(fmt.Sprintf("%f", original))

	// Act
	_, err := service.Withdraw(context.Background(), userID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

func TestWalletService_Transfer(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

	fromUserID := 1
	toUserID := 2
	amount := 30.00
	fromNewBalance := 20.00
	toNewBalance := 50.00

	mockSQL.ExpectBegin()

	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, fromUserID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(fromNewBalance))
	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", fromUserID)).SetVal(fmt.Sprintf("%f", fromNewBalance+amount))

	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance \+ \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, toUserID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(toNewBalance))
	mockSQL.ExpectExec(`INSERT INTO transactions \(from_user_id, to_user_id, amount, transaction_type\)`).
		WithArgs(fromUserID, toUserID, amount, "transfer").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", toUserID)).SetVal(fmt.Sprintf("%f", toNewBalance-amount))
	mockSQL.ExpectCommit()
	mockRedis.ExpectSet(`wallet_balance:1`, fromNewBalance, 0).SetVal("OK")
	mockRedis.ExpectSet(`wallet_balance:2`, toNewBalance, 0).SetVal("OK")

	// Act
	fromBalance, toBalance, err := service.Transfer(context.Background(), fromUserID, toUserID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

func TestWalletService_Transfer_Error(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()

	fromUserID := 1
	toUserID := 2
	amount := 30.00
	fromNewBalance := 20.00
	toNewBalance := 50.00

	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", fromUserID)).SetVal(fmt.Sprintf("%f", fromNewBalance+amount))
	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", toUserID)).SetVal(fmt.Sprintf("%f", toNewBalance-amount))

	mockSQL.ExpectBegin()
	mockSQL.ExpectQuery(`UPDATE wallets SET balance = balance - \$1 WHERE user_id = \$2 RETURNING balance`).
		WithArgs(amount, fromUserID).
		WillReturnError(sql.ErrConnDone)
	mockSQL.ExpectRollback()

	// Act
	_, _, err := service.Transfer(context.Background(), fromUserID, toUserID, amount)

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

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet redis expectations: %v", err)
	}
}

// GetBalance
func TestWalletService_GetBalance_FromRedis(t *testing.T) {
	service, _, mockRedis := setupMockRepo()
	userID := 1
	balance := 150.00

	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", userID)).SetVal(fmt.Sprintf("%f", balance))

	returnedBalance, err := service.GetBalance(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance != returnedBalance {
		t.Fatalf("expected to user balance to be %v, got %v", returnedBalance, balance)
	}
}

func TestWalletService_GetBalance_FromDatabase(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()
	userID := 1
	balance := 150.00

	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", userID)).RedisNil()

	mockSQL.ExpectQuery(`SELECT balance FROM wallets WHERE user_id=\$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(balance))
	mockRedis.ExpectSet(fmt.Sprintf("wallet_balance:%d", userID), balance, 0).SetVal("OK")

	returnedBalance, err := service.GetBalance(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance != returnedBalance {
		t.Fatalf("expected to user balance to be %v, got %v", returnedBalance, balance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet SQL expectations: %v", err)
	}

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet Redis expectations: %v", err)
	}
}

func TestWalletService_GetBalance_FromDatabase_FailedToSetCache(t *testing.T) {
	// Arrange
	service, mockSQL, mockRedis := setupMockRepo()
	userID := 1
	balance := 150.00

	mockRedis.ExpectGet(fmt.Sprintf("wallet_balance:%d", userID)).RedisNil()
	mockSQL.ExpectQuery(`SELECT balance FROM wallets WHERE user_id=\$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(balance))

	mockRedis.ExpectSet(fmt.Sprintf("wallet_balance:%d", userID), balance, 0).SetErr(errors.New("failed to set cache"))

	returnedBalance, err := service.GetBalance(context.Background(), userID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if returnedBalance != 0.0 {
		t.Fatalf("expected to user balance to be %v, got %v", 0, returnedBalance)
	}

	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet SQL expectations: %v", err)
	}

	if err := mockRedis.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet Redis expectations: %v", err)
	}
}
