package logs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func initLogger() {
	logger = logrus.New()

	formatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			funcRelativePathIndex := strings.LastIndex(f.Function, ".") + 1
			funcName := f.Function[funcRelativePathIndex:]
			return funcName, fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}

	logger.SetReportCaller(true)
	logger.SetFormatter(formatter)
	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  "./logs/info.log",
		logrus.WarnLevel:  "./logs/warn.log",
		logrus.ErrorLevel: "./logs/error.log",
		logrus.FatalLevel: "./logs/error.log",
		logrus.PanicLevel: "./logs/error.log",
	}
	logger.Hooks.Add(lfshook.NewHook(
		pathMap,
		formatter,
	))
	logger.WriterLevel(logrus.InfoLevel)
}

func GetLogger() *logrus.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
