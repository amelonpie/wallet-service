package endpoint

import (
	"github.com/amelonpie/wallet-service/internal/wallet"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRouters(router *gin.Engine) {
	var repo wallet.Repository

	var err error

	repo, err = wallet.InitRepository()
	if err != nil {
		msg := "failed to initialize wallet repository"
		logrus.Fatalf("%s: %v", msg, err)

		panic(err)
	}

	var svc wallet.Service
	svc, err = wallet.InitService(repo)

	if err != nil {
		msg := "failed to initialize wallet service"
		logrus.Fatalf("%s: %v", msg, err)

		panic(err)
	}

	ep := newEndpoint(svc)

	wallet := router.Group("/wallet")
	{
		addTransactionRoutes(wallet, ep)
		addViewRoutes(wallet, ep)
	}
}
