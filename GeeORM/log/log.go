package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

/**
 *	支持日志分级（Info、Error、Disabled 三级）。
 *	不同层级日志显示时使用不同的颜色区分。
 *	显示打印日志代码对应的文件名和行号。
 *	[info ] 蓝色；[error] 红色
 *	log.Lshortfile 支持显示文件名和代码行号
 *	log.LstdFlags 标准 log 配置 Ldate | Ltime
 */
var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// log methods
var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

// log levels
const (
	infoLevel = iota
	ErrorLevel
	Disabled
)

/**
 * 通过控制 log 等级阈值，来控制日志是否打印
 * 例如设置了 level = ErrorLevel，那么 info 就不会被打印
 */
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if infoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
