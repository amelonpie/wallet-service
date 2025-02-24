package endpoints

import (
	"errors"
	"net/http"

	"github.com/amelonpie/wallet-service/internal/wallet"
	"github.com/amelonpie/wallet-service/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Endpoint struct {
	Logger *logrus.Entry
	Svc    *wallet.Service
}

func newEndpoint(svc wallet.Service) *Endpoint {
	logger := log.NewLogger("endpoint").WithField("module", "endpoints")

	return &Endpoint{
		Logger: logger,
		Svc:    &svc,
	}
}

func epInstance(c *gin.Context) (*Endpoint, error) {
	epAny, exists := c.Get("endpoint")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Endpoint not found in context"})
		//nolint:err113 // no need to define error class
		return nil, errors.New("endpoint not found")
	}

	ep, ok := epAny.(*Endpoint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast Endpoint"})
		//nolint:err113 // no need to define error class
		return nil, errors.New("failed to cast endpoint")
	}

	return ep, nil
}

func epLogger(c *gin.Context) (*logrus.Entry, error) {
	ep, err := epInstance(c)
	if err != nil {
		logrus.Fatal("fail to get endpoint")

		return nil, err
	}

	return ep.Logger, nil
}

func epSvc(c *gin.Context) (*wallet.Service, error) {
	ep, err := epInstance(c)
	if err != nil {
		logrus.Fatal("fail to get endpoint")
		return nil, err
	}

	return ep.Svc, nil
}
