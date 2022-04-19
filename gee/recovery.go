package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(message string) string {
	var pcs [32]uintptr
	// Callers 用来返回调用栈的程序计数器, 第 0 个 Caller 是 Callers 本身，
	// 第 1 个是上一层 trace，第 2 个是再上一层的 defer func。
	// 因此，为了日志简洁一点，我们跳过了前 3 个 Caller。
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		// 获取函数
		fn := runtime.FuncForPC(pc)
		// 获取调用该函数的文件名和行号
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		// defer recover 机制只能针对于当前函数
		// 以及当前函数直接调用的函数的 panic
		// 如果没有 c.Next()，则 handler 不是 Recovery
		// 直接调用的函数，无法 recover，
		// panic 会被 net/http 自带的 recover 机制捕获
		c.Next()
	}
}
