package kardLog

import (
	"kard/src/global/variable"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CreateZapFactory(entry func(zapcore.Entry) error) *zap.Logger {
	// 获取程序所处的模式：  开发调试 、 生产
	//variable.WebYml := yml_config.CreateYamlFactory()

	encoderConfig := zap.NewProductionEncoderConfig()

	timePrecision := variable.WebYml.GetString("Logs.TimePrecision")
	var recordTimeFormat string
	switch timePrecision {
	case "second":
		recordTimeFormat = "2006-01-02 15:04:05"
	case "millisecond":
		recordTimeFormat = "2006-01-02 15:04:05.000"
	default:
		recordTimeFormat = "2006-01-02 15:04:05"

	}
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(recordTimeFormat))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "created_at" // 生成json格式日志的时间键字段，默认为 ts,修改以后方便日志导入到 ELK 服务器

	var encoder zapcore.Encoder
	switch variable.WebYml.GetString("Logs.TextFormat") {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // 普通模式
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig) // json格式
	default:
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // 普通模式
	}

	//写入器
	fileName := variable.BasePath + variable.WebYml.GetString("Logs.GoSkeletonLogName")
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName,                                  //日志文件的位置
		MaxSize:    variable.WebYml.GetInt("Logs.MaxSize"),    //在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups: variable.WebYml.GetInt("Logs.MaxBackups"), //保留旧文件的最大个数
		MaxAge:     variable.WebYml.GetInt("Logs.MaxAge"),     //保留旧文件的最大天数
		Compress:   variable.WebYml.GetBool("Logs.Compress"),  //是否压缩/归档旧文件
	}
	writer := zapcore.AddSync(lumberJackLogger)
	// 开始初始化zap日志核心参数，
	//参数一：编码器
	//参数二：写入器
	//参数三：参数级别，debug级别支持后续调用的所有函数写日志，如果是 fatal 高级别，则级别>=fatal 才可以写日志

	var allCore []zapcore.Core
	zapCore := zapcore.NewCore(encoder, writer, zap.InfoLevel)
	allCore = append(allCore, zapCore)

	// 以下是 调试（生产）模式所需要的代码
	appDebug := variable.WebYml.GetBool("AppDebug")
	// 判断程序当前所处的模式，调试模式直接返回一个便捷的zap日志管理器地址，所有的日志打印到控制台
	if appDebug {
		consoleDebugging := zapcore.AddSync(os.Stdout)
		allCore = append(allCore, zapcore.NewCore(encoder, consoleDebugging, zap.DebugLevel))

		// if logger, err := zap.NewDevelopment(zap.Hooks(entry)); err == nil {
		// 	return logger
		// } else {
		// 	log.Fatal("创建zap日志包失败，详情：" + err.Error())
		// }
	}

	core := zapcore.NewTee(allCore...)
	return zap.New(core, zap.AddCaller(), zap.Hooks(entry))
}

// GoSkeleton 系统运行日志钩子函数
// 1.单条日志就是一个结构体格式，本函数拦截每一条日志，您可以进行后续处理，例如：推送到阿里云日志管理面板、ElasticSearch 日志库等

func ZapLogHandler(entry zapcore.Entry) error {

	// 参数 entry 介绍
	// entry  参数就是单条日志结构体，主要包括字段如下：
	//Level      日志等级
	//Time       当前时间
	//LoggerName  日志名称
	//Message    日志内容
	//Caller     各个文件调用路径
	//Stack      代码调用栈

	//这里启动一个协程，hook丝毫不会影响程序性能，
	go func(paramEntry zapcore.Entry) {
		//fmt.Println(" GoSkeleton  hook ....，你可以在这里继续处理系统日志....")
		//fmt.Printf("%#+v\n", paramEntry)
	}(entry)
	return nil
}
