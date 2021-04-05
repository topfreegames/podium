// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//CM is a Checked Message like
type CM interface {
	Write(fields ...zap.Field)
}

//D is a debug logger
func D(logger *zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.DebugLevel, message, callback...)
}

//I is a info logger
func I(logger *zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.InfoLevel, message, callback...)
}

//W is a warn logger
func W(logger *zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.WarnLevel, message, callback...)
}

//E is a error logger
func E(logger *zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.ErrorLevel, message, callback...)
}

//P is a panic logger
func P(logger *zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.PanicLevel, message, callback...)
}

func defaultWrite(l CM) {
	l.Write()
}

func log(logger *zap.Logger, logLevel zapcore.Level, message string, callback ...func(l CM)) {
	cb := defaultWrite
	if len(callback) == 1 {
		cb = callback[0]
	}
	if cm := logger.Check(logLevel, message); cm != nil {
		cb(cm)
	}
}

//LoggerOptions is a struct to manipulate Logger role
type LoggerOptions struct {
	zapcore.WriteSyncer
	RemoveTimestamp bool
}

//CreateLoggerWithLevel instantiate a new logger with level
func CreateLoggerWithLevel(ll zapcore.Level, options LoggerOptions) *zap.Logger {
	atom := zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
	if options.RemoveTimestamp {
		encoderCfg.TimeKey = ""
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(options.WriteSyncer),
		atom,
	))

	atom.SetLevel(ll)

	return logger
}
