package endpoint

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func addViewRoutes(wallet *gin.RouterGroup, ep *Endpoint) {
	wallet.GET("/wallet/:user_id/balance", func(c *gin.Context) {
		c.Set("endpoint", ep)
		balanceHandler(c)
	})
	wallet.GET("/wallet/:user_id/transactions", func(c *gin.Context) {
		c.Set("endpoint", ep)
		transactionsHandler(c)
	})
}

func balanceHandler(c *gin.Context) {
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

	svc, _ := epSvc(c)

	balance, err := (*svc).GetBalance(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		endpointLogger.WithFields(logrus.Fields{
			"err":     err,
			"user_id": userID,
		}).Errorf("failed to get balance")

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
	endpointLogger.WithFields(logrus.Fields{
		"user_id": userID,
		"balance": balance,
	}).Info("successful get balance")
}

func transactionsHandler(c *gin.Context) {
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

	svc, _ := epSvc(c)

	history, err := (*svc).GetTransactionHistory(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
	endpointLogger.WithFields(logrus.Fields{
		"user_id":     userID,
		"transaction": history,
	}).Info("successful get transaction history")
}
