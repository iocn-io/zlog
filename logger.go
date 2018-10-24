package zlog

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"strings"
)

type logConfig struct {
	Desc      string
	Level     string
	Stdout    bool
	Encoding  string
	AddCaller bool
	Color     bool
	FilesOut  bool
	LogsPath  []*logFilePath
}

type logFilePath struct {
	Level string
	Hook  *lumberjack.Logger
}

var (
	Zlog *zap.Logger
)

func init() {
	Zlog = New()
}

func New() *zap.Logger {
	logcfg := logConfig{
		Desc:      "development",
		Level:     "debug",
		Stdout:    true,
		Encoding:  "console",
		AddCaller: true,
		Color:     true,
		FilesOut:  false,
		LogsPath: []*logFilePath{{
			Level: "debug",
			Hook: &lumberjack.Logger{
				Filename:   "./logs/gaven.log", // Filename is the file to write logs to.
				MaxSize:    1024,               // megabytes
				MaxAge:     7,                  // days
				MaxBackups: 3,                  // the maximum number of old log files to retain.
			},
		},
		},
	}
	exists, err := pathExists("./logs.json")
	if err != nil {
		fmt.Println(err)
	}
	if exists {
		file, err := ioutil.ReadFile("./logs.json")
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(file, &logcfg); err != nil {
			panic(err)
		}
	}
	// Output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)
	var encoderConfig zapcore.EncoderConfig
	var fileEncoder, consoleEncoder zapcore.Encoder
	if strings.EqualFold(logcfg.Desc, "production") {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else if strings.EqualFold(logcfg.Desc, "development") {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		fmt.Println("'" + logcfg.Desc + "' in the configuration file for desc is an invalid value, it could be 'development' or 'production'")
		os.Exit(1)
	}
	if strings.EqualFold(logcfg.Encoding, "json") {
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else if strings.EqualFold(logcfg.Encoding, "console") {
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		if logcfg.Color {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		fmt.Println("'" + logcfg.Encoding + "' in the configuration file for encoding is an invalid value, it could be 'json' or 'console'")
		os.Exit(1)
	}
	var cores []zapcore.Core
	if !logcfg.Stdout && !logcfg.FilesOut || logcfg.Stdout {
		cores = append(cores, zapcore.NewCore(consoleEncoder, consoleDebugging, getLevel(logcfg.Level)))
	}
	if logcfg.FilesOut {
		if len(logcfg.LogsPath) > 0 {
			for i := 0; i < len(logcfg.LogsPath); i++ {
				cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(logcfg.LogsPath[i].Hook), getLevel(logcfg.LogsPath[i].Level)))
			}
		}
	}
	core := zapcore.NewTee(cores...)
	// From a zapcore.Core to construct a Logger.
	var logger *zap.Logger
	if logcfg.AddCaller {
		logger = zap.New(core, zap.AddCaller())
	} else {
		logger = zap.New(core)
	}
	logger.Info("Default Logger init success")
	return logger
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "panic", "dpanic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.DebugLevel
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
