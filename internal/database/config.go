package database

import (
	"github.com/amelonpie/wallet-service/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	PostgreAddr string
	RedisAddr   string
	RedisPwd    string
	RedisDB     int
	logger      *logrus.Entry
}

func NewDatabaseConfig() *Config {
	return &Config{
		PostgreAddr: viper.GetString("postgresql.address"),
		RedisAddr:   viper.GetString("redis.address"),
		RedisPwd:    viper.GetString("redis.password"),
		RedisDB:     viper.GetInt("redis.db"),
		logger:      log.NewLogger("database").WithField("module", "database"),
	}
}
