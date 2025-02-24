package endpoints

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amelonpie/wallet-service/internal/wallet"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

/*
// why linter discourage this order but have to place all of them alphabet
import (
	// standard
	"context"
	"net/http"
	"strconv"

	// external
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	// internal
	"github.com/amelonpie/wallet-service/internal/wallet"
)
*/

func addTransactionRoutes(wallet *gin.RouterGroup, ep *Endpoint) {
	wallet.POST("/:user_id/deposit", func(c *gin.Context) {
		c.Set("endpoint", ep)
		depositHandler(c)
	})
	wallet.POST("/:user_id/withdraw", func(c *gin.Context) {
		c.Set("endpoint", ep)
		withdrawHandler(c)
	})
	wallet.POST("/transfer", func(c *gin.Context) {
		c.Set("endpoint", ep)
		transferHandler(c)
	})
}

func depositHandler(c *gin.Context) {
	handleTransactionRequest(c, "deposit")
}

func withdrawHandler(c *gin.Context) {
	handleTransactionRequest(c, "withdraw")
}

//nolint:funlen
func handleTransactionRequest(c *gin.Context, requestType string) {
	endpointLogger, err := epLogger(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
		reqBody, _ := c.GetRawData()
		endpointLogger.WithFields(logrus.Fields{
			"err":          "server internal error",
			"request_body": reqBody,
		}).Error("failed to get logger")

		return
	}

	// Parse user_id from path
	userIDParam := c.Param("user_id")
	userID, err := strconv.Atoi(userIDParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		endpointLogger.WithFields(logrus.Fields{
			"err":           err,
			"user_id_param": userIDParam,
		}).Error("invalid user_id")

		return
	}

	// Parse JSON request body
	var req any

	var serviceMethod func(context.Context, int, float64) (float64, error)

	var svc *wallet.Service

	svc, _ = epSvc(c)

	switch requestType {
	case "deposit":
		req = &DepositRequest{}
		serviceMethod = (*svc).Deposit
	case "withdraw":
		req = &WithdrawRequest{}
		serviceMethod = (*svc).Withdraw
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request type"})
		reqBody, _ := c.GetRawData()
		endpointLogger.WithFields(logrus.Fields{
			"err":          err,
			"request_body": reqBody,
		}).Error("invalid request type")

		return
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		reqBody, _ := c.GetRawData()
		endpointLogger.WithFields(logrus.Fields{
			"err":          err,
			"request_body": reqBody,
		}).Error("invalid request body")

		return
	}

	// Call the service method
	var amount float64

	if requestType == "deposit" {
		depositReq, ok := req.(*DepositRequest)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
			reqBody, _ := c.GetRawData()
			endpointLogger.WithFields(logrus.Fields{
				"err":          "server internal error",
				"request_body": reqBody,
			}).Error("server internal error")

			return
		}

		amount = depositReq.Amount
	} else if requestType == "withdraw" {
		withdrawReq, ok := req.(*WithdrawRequest)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
			reqBody, _ := c.GetRawData()
			endpointLogger.WithFields(logrus.Fields{
				"err":          "server internal error",
				"request_body": reqBody,
			}).Error("server internal error")

			return
		}

		amount = withdrawReq.Amount
	}

	newBalance, err := serviceMethod(context.Background(), userID, amount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		endpointLogger.WithFields(logrus.Fields{
			"err":     err,
			"user_id": userID,
			"amount":  amount,
		}).Errorf("failed to %s", requestType)

		return
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"new_balance": newBalance,
	})
	endpointLogger.WithFields(logrus.Fields{
		"user_id":    userID,
		"amount":     amount,
		"newBalance": newBalance,
	}).Infof("successful %s", requestType)
}

func transferHandler(c *gin.Context) {
	endpointLogger, err := epLogger(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server internal error"})
		reqBody, _ := c.GetRawData()
		endpointLogger.WithFields(logrus.Fields{
			"err":          "server internal error",
			"request_body": reqBody,
		}).Error("failed to get logger")

		return
	}

	var req TransferRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		reqBody, _ := c.GetRawData()
		endpointLogger.WithFields(logrus.Fields{
			"err":          err,
			"request_body": reqBody,
		}).Error("invalid request body")

		return
	}

	svc, _ := epSvc(c)

	newFromBalance, newToBalance, err := (*svc).Transfer(
		context.Background(),
		req.FromUserID,
		req.ToUserID,
		req.Amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"from_balance": newFromBalance})

	endpointLogger.WithFields(logrus.Fields{
		"from_user_id":     req.FromUserID,
		"to_user_id":       req.ToUserID,
		"amount":           req.Amount,
		"new_from_balance": newFromBalance,
		"new_to_balance":   newToBalance,
	}).Infof("successful transfer")
}
