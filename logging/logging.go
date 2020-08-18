package logging

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Conf logging conf
type Conf struct {
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	FileName   string
}

// New create a logger
func New(conf Conf) *zap.SugaredLogger {
	var writeSyncer zapcore.WriteSyncer

	writeSyncer = getLogWriter(conf)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	sugarlogger := logger.Sugar()
	return sugarlogger
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(conf Conf) zapcore.WriteSyncer {
	if len(conf.FileName) == 0 {
		return zapcore.AddSync(os.Stdout)
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf.FileName,
		MaxSize:    conf.MaxSize,
		MaxBackups: conf.MaxBackups,
		MaxAge:     conf.MaxAge,
		Compress:   conf.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}
