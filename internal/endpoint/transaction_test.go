package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amelonpie/wallet-service/internal/wallet"
	"github.com/gin-gonic/gin"
)

// func TestMain(m *testing.M) {
// 	goleak.VerifyTestMain(m)
// }

type mockWalletService struct {
	DepositFunc               func(ctx context.Context, userID int, amount float64) (float64, error)
	WithdrawFunc              func(ctx context.Context, userID int, amount float64) (float64, error)
	TransferFunc              func(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error)
	GetBalanceFunc            func(ctx context.Context, userID int) (float64, error)
	GetTransactionHistoryFunc func(ctx context.Context, userID int) ([]wallet.Transaction, error)
}

func (m *mockWalletService) Deposit(ctx context.Context, userID int, amount float64) (float64, error) {
	return m.DepositFunc(ctx, userID, amount)
}
func (m *mockWalletService) Withdraw(ctx context.Context, userID int, amount float64) (float64, error) {
	return m.WithdrawFunc(ctx, userID, amount)
}
func (m *mockWalletService) Transfer(ctx context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error) {
	return m.TransferFunc(ctx, fromUserID, toUserID, amount)
}
func (m *mockWalletService) GetBalance(ctx context.Context, userID int) (float64, error) {
	return m.GetBalanceFunc(ctx, userID)
}
func (m *mockWalletService) GetTransactionHistory(ctx context.Context, userID int) ([]wallet.Transaction, error) {
	return m.GetTransactionHistoryFunc(ctx, userID)
}

var _ wallet.Service = (*mockWalletService)(nil)

func TestDepositHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{
		DepositFunc: func(_ context.Context, userID int, amount float64) (float64, error) {
			if userID == 0 {
				return 0, errors.New("invalid user_id for deposit")
			}

			return 100.0 + amount, nil
		},
	}

	ep := newEndpoint(mockSvc)

	rg := router.Group("/wallet")
	rg.POST("/:user_id/deposit", func(c *gin.Context) {
		c.Set("endpoint", ep)
		depositHandler(c)
	})

	t.Run("valid deposit", func(t *testing.T) {
		body, _ := json.Marshal(DepositRequest{Amount: 50.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/123/deposit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %v, got %v", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid user_id (not integer)", func(t *testing.T) {
		body, _ := json.Marshal(DepositRequest{Amount: 50.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/abc/deposit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		body := []byte("invalid json")
		req, _ := http.NewRequest(http.MethodPost, "/wallet/123/deposit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("deposit service error", func(t *testing.T) {
		// let server return error when user_id == 0
		body, _ := json.Marshal(DepositRequest{Amount: 50.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/0/deposit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestWithdrawHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{
		WithdrawFunc: func(_ context.Context, userID int, amount float64) (float64, error) {
			if userID < 0 {
				return 0, errors.New("invalid user_id for withdraw")
			}
			if amount > 100.0 {
				return 0, errors.New("insufficient balance")
			}
			return 100.0 - amount, nil
		},
	}

	ep := newEndpoint(mockSvc)
	rg := router.Group("/wallet")
	rg.POST("/:user_id/withdraw", func(c *gin.Context) {
		c.Set("endpoint", ep)
		withdrawHandler(c)
	})

	t.Run("valid withdraw", func(t *testing.T) {
		body, _ := json.Marshal(WithdrawRequest{Amount: 30.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/123/withdraw", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %v, got %v", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid user_id (not integer)", func(t *testing.T) {
		body, _ := json.Marshal(WithdrawRequest{Amount: 30.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/xyz/withdraw", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		body := []byte("invalid json")
		req, _ := http.NewRequest(http.MethodPost, "/wallet/123/withdraw", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("withdraw service error - insufficient funds", func(t *testing.T) {
		body, _ := json.Marshal(WithdrawRequest{Amount: 999.0})
		req, _ := http.NewRequest(http.MethodPost, "/wallet/123/withdraw", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})
}

func TestHandleTransactionRequest_InvalidRequestType(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{}
	ep := newEndpoint(mockSvc)

	router.POST("/transaction/:user_id", func(c *gin.Context) {
		c.Set("endpoint", ep)
		handleTransactionRequest(c, "other")
	})

	// Act
	reqBody := `{"amount": 100}`
	req := httptest.NewRequest(http.MethodPost, "/transaction/123", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %v, got %v", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if response["error"] != "invalid request type" {
		t.Errorf("expected error message to be 'invalid request type', got %v", response["error"])
	}
}

func TestTransferHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{
		TransferFunc: func(_ context.Context, fromUserID, toUserID int, amount float64) (float64, float64, error) {
			if fromUserID < 0 || toUserID < 0 {
				return 0, 0, errors.New("invalid user IDs")
			}
			if amount > 100 {
				return 0, 0, errors.New("insufficient balance")
			}
			return 50.0, 150.0, nil
		},
	}

	ep := newEndpoint(mockSvc)

	router.POST("/transfer", func(c *gin.Context) {
		c.Set("endpoint", ep)
		transferHandler(c)
	})

	t.Run("valid transfer", func(t *testing.T) {
		body, _ := json.Marshal(TransferRequest{FromUserID: 1, ToUserID: 2, Amount: 20.0})
		req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected %v, got %v", http.StatusOK, w.Code)
		}
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		body := []byte("invalid json")
		req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("transfer service error", func(t *testing.T) {
		body, _ := json.Marshal(TransferRequest{FromUserID: -1, ToUserID: 2, Amount: 50.0})
		req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})
}
