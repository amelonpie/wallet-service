package endpoints

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amelonpie/wallet-service/internal/wallet"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBalanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{
		GetBalanceFunc: func(_ context.Context, userID int) (float64, error) {
			if userID < 1 {
				return 0, errors.New("user not found")
			}
			return 999.99, nil
		},
	}
	ep := newEndpoint(mockSvc)
	rg := router.Group("/wallet")
	rg.GET("/wallet/:user_id/balance", func(c *gin.Context) {
		c.Set("endpoint", ep)
		balanceHandler(c)
	})

	t.Run("valid balance request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/123/balance", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid user_id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/abc/balance", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/0/balance", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTransactionsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockSvc := &mockWalletService{
		GetTransactionHistoryFunc: func(_ context.Context, userID int) ([]wallet.Transaction, error) {
			if userID == 0 {
				return nil, errors.New("no transactions found for user 0")
			}
			return []wallet.Transaction{
				{TransactionID: 1, FromUserID: userID, Amount: 50, TransactionType: "deposit"},
				{TransactionID: 2, ToUserID: userID, Amount: -20, TransactionType: "withdraw"},
			}, nil
		},
	}
	ep := newEndpoint(mockSvc)
	rg := router.Group("/wallet")
	rg.GET("/wallet/:user_id/transactions", func(c *gin.Context) {
		c.Set("endpoint", ep)
		transactionsHandler(c)
	})

	t.Run("valid user_id", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/123/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid user_id format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/abc/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/wallet/wallet/0/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
