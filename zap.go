package main

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

//var logger *zap.Logger
//var sugarLogger *zap.SugaredLogger
//
//func main() {
//	InitLogger()
//	defer logger.Sync()
//	simpleHttpGet("www.google.com")
//	simpleHttpGet("http://www.google.com")
//}
//
//func InitLogger() {
//	logger, _ = zap.NewProduction()
//}
//
//func simpleHttpGet(url string) {
//	resp, err := http.Get(url)
//	if err != nil {
//		logger.Error(
//			"Error fetching url..",
//			zap.String("url", url),
//			zap.Error(err))
//	} else {
//		logger.Info("Success..",
//			zap.String("statusCode", resp.Status),
//			zap.String("url", url))
//		resp.Body.Close()
//	}
//}
var sugarLogger *zap.SugaredLogger

func main() {

	InitLogger()
	defer sugarLogger.Sync()
	simpleHttpGet("www.baidu.com")
	simpleHttpGet("http://www.baidu.com")
}

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller()) //zap.AddCaller()  在日志中添加调用函数
	sugarLogger = logger.Sugar()
}

func simpleHttpGet(url string) {
	sugarLogger.Debugf("Trying to hit GET request for %s", url)
	resp, err := http.Get(url)
	if err != nil {
		sugarLogger.Errorf("Error fetching URL %s : Error = %s", url, err)
	} else {
		sugarLogger.Infof("Success! statusCode = %s for URL %s", resp.Status, url)
		resp.Body.Close()
	}
}
func getEncoder() zapcore.Encoder {
	//return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())  用JOSN编码写日志
	encodeCofig := zap.NewProductionEncoderConfig()
	encodeCofig.EncodeTime = zapcore.ISO8601TimeEncoder                //将日志时间从时间戳（默认）修改为form形式
	return zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()) //普通日志
}

//直接打开文件
//func getLogWriter() zapcore.WriteSyncer {
//	file, _ := os.Create("./test.log") //每次打开新文件写入日志
//	//file, _ :=os.OpenFile("./test.log",os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)  //追加写日志
//	return zapcore.AddSync(file)
//}
//日志切割
func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./test.log",
		MaxSize:    10,    //MB
		MaxBackups: 5,     //备份数
		MaxAge:     30,    //备份天数
		Compress:   false, //是否压缩
	}
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
