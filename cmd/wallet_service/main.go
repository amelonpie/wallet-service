package main

import (
	endpoints "github.com/amelonpie/wallet-service/internal/endpoint"
	"github.com/amelonpie/wallet-service/pkg/config"
	"github.com/amelonpie/wallet-service/pkg/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initRouter(mainLogger *logrus.Logger) {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Cache-Control", "Content-Type"},
		AllowCredentials: true,
	}))

	endpoints.SetupRouters(router)

	if err := router.Run(":" + viper.GetString("api_port")); err != nil {
		mainLogger.Error(err)
	}
}

func main() {
	err := config.Initialize()
	if err != nil {
		logrus.WithField("err", err).Fatalf("failed to initialize config")
		return
	}

	var mainLogger *logrus.Logger
	mainLogger, err = log.Initialize()
	if err != nil {
		logrus.WithField("err", err).Fatalf("failed to initialize logger")
		return
	}

	mainLogger.Infoln("wallet service api start running")
	initRouter(mainLogger)
}
