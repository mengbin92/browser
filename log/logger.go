package log

import (
	"os"

	"github.com/mengbin92/browser/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultLogger,stdout
func DefaultLogger(logConf *conf.Log) *zap.Logger {
	var coreArr []zapcore.Core

	//获取编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        //指定时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder //按级别显示不同颜色
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder      //显示完整文件路径

	var encoder zapcore.Encoder //NewJSONEncoder()输出json格式，NewConsoleEncoder()输出普通文本格式
	if logConf.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	//日志级别
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //error级别
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //info和debug级别,debug级别是最低的
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	infoCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), lowPriority)   //第三个及之后的参数为写入文件的日志级别,ErrorLevel模式只记录error级别的日志
	errorCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), highPriority) //第三个及之后的参数为写入文件的日志级别,ErrorLevel模式只记录error级别的日志

	coreArr = append(coreArr, infoCore)
	coreArr = append(coreArr, errorCore)

	return setLogLevel(zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()), logConf.GetLevel())
}

func setLogLevel(log *zap.Logger, level int32) *zap.Logger {
	switch level {
	case -1:
		return log.WithOptions(zap.IncreaseLevel(zapcore.DebugLevel))
	case 0:
		return log.WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))
	case 1:
		return log.WithOptions(zap.IncreaseLevel(zapcore.WarnLevel))
	case 3:
		return log.WithOptions(zap.IncreaseLevel(zapcore.DPanicLevel))
	case 4:
		return log.WithOptions(zap.IncreaseLevel(zapcore.PanicLevel))
	case 5:
		return log.WithOptions(zap.IncreaseLevel(zapcore.FatalLevel))
	default:
		return log.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))
	}
}
