package config

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var confPath string

// Initialize inits viper app settings
func Initialize() error {

	flag.StringVar(&confPath, "c", "", "configuration file path")

	if !flag.Parsed() {
		flag.Parse()
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("log_path", "/var/log")
	viper.SetDefault("app_name", "main")

	if confPath != "" {
		viper.SetConfigFile(confPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file:", viper.ConfigFileUsed())
	}

	viper.Debug()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
	})

	return nil
}
