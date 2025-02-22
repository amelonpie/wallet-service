package log

import (
	"os"
	"path"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	logBasePath  string
	mainLogger   *logrus.Logger
	mainLogPath  string
	warnLogPath  string
	errorLogPath string
	logLevel     logrus.Level = logrus.DebugLevel
)

// Initialize init logger and returns main logger
func Initialize() (*logrus.Logger, error) {
	if mainLogger != nil {
		return mainLogger, nil
	}

	logBasePath = viper.GetString("log_path")
	logrus.Info("Setting logBasePath: ", logBasePath)

	appName := viper.GetString("app_name")
	mainLogPath = path.Join(logBasePath, appName+".log")
	warnLogPath = path.Join(logBasePath, appName+"_warn.log")
	errorLogPath = path.Join(logBasePath, appName+"_error.log")

	if _, err := os.Stat(logBasePath); os.IsNotExist(err) {
		if err := os.MkdirAll(logBasePath, 0755); err != nil {
			logrus.WithField("err", err).Error("failed to create log directory")

			return nil, err
		}
		logrus.Info("created log directory:", logBasePath)
	}

	mainLogger = logrus.New()
	mainLogger.Out = os.Stdout
	mainLogger.SetFormatter(&logrus.JSONFormatter{})
	mainLogger.SetReportCaller(true)

	var err error
	logLevel, err = logrus.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		logrus.WithField("err", err).Warn("invalid log level")
		logLevel = logrus.InfoLevel
	}
	mainLogger.SetLevel(logLevel)

	mainLogger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.DebugLevel: mainLogPath,
		logrus.InfoLevel:  mainLogPath,
		logrus.WarnLevel:  mainLogPath,
		logrus.ErrorLevel: mainLogPath,
	}, &logrus.JSONFormatter{}))
	mainLogger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.WarnLevel:  warnLogPath,
		logrus.ErrorLevel: warnLogPath,
	}, &logrus.TextFormatter{}))
	mainLogger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.ErrorLevel: errorLogPath,
	}, &logrus.TextFormatter{}))

	return mainLogger, nil
}

// GetLogBasePath returns log base path
func GetLogBasePath() string {
	return logBasePath
}

// MainLogger returns main logger
func MainLogger() *logrus.Logger {
	return mainLogger
}

// NewLogger creates new logger, logs hooked to main logger
func NewLogger(filename string) *logrus.Logger {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetReportCaller(true)

	logger.SetLevel(logLevel)

	loggerPath := path.Join(logBasePath, viper.GetString("app_name")+"_"+filename+".log")
	logger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.DebugLevel: loggerPath,
		logrus.InfoLevel:  loggerPath,
		logrus.WarnLevel:  loggerPath,
		logrus.ErrorLevel: loggerPath,
	}, &logrus.JSONFormatter{}))
	logger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.DebugLevel: mainLogPath,
		logrus.InfoLevel:  mainLogPath,
		logrus.WarnLevel:  mainLogPath,
		logrus.ErrorLevel: mainLogPath,
	}, &logrus.JSONFormatter{}))
	logger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.WarnLevel:  warnLogPath,
		logrus.ErrorLevel: warnLogPath,
	}, &logrus.TextFormatter{}))
	logger.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.ErrorLevel: errorLogPath,
	}, &logrus.TextFormatter{}))

	return logger
}

// NewSoloLogger creates solo file logger at logBasePath
func NewSoloLogger(filename string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	loggerPath := path.Join(logBasePath, viper.GetString("app_name")+"_"+filename+".log")
	file, err := os.OpenFile(loggerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	return logger
}
