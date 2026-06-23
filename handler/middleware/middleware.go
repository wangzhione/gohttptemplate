// Package middleware 拦截器主要用于所有请求, 注入 trace id, 统一处理 panic
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/wangzhione/gohttptemplate/configs"
	"github.com/wangzhione/sbp/chain"
)

// ResponseWriterPanicError 自定义 589 类似的 5xx 当成 panic error
func ResponseWriterPanicError(w http.ResponseWriter) {
	w.WriteHeader(589)
	fmt.Fprintf(w, `{"code":"589", "message":"Internal Server Panic Error"}`)
}

// ispprofpath 检查请求路径是否为 pprof 调试路由
func ispprofpath(path string) bool {
	return path == "/debug/pprof" || strings.HasPrefix(path, "/debug/pprof/")
}

// token 通过 header "pprofbearer: Bearer <token>" 传递
func verifypprofbearer(r *http.Request, pprofbearer string) bool {
	if !ispprofpath(r.URL.Path) {
		return true
	}

	pprofheader := r.Header.Get("pprofbearer")
	if pprofheader == "" {
		return false
	}

	// 解析 Bearer token 格式: "Bearer <token>"
	const bearerprefix = "Bearer "
	if len(pprofheader) > len(bearerprefix) && pprofheader[:len(bearerprefix)] == bearerprefix {
		return pprofheader[len(bearerprefix):] == pprofbearer
	}
	return false
}

// ResponseWriterPprofUnauthorized 返回 pprof 未授权响应
func ResponseWriterPprofUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"code":"401", "message":"Pprof Access Token Required or Invalid"}`)
}

// MainMiddleware 拦截器, 默认 serve 所以拦截器集中在这里
// 自动从 configs.G.Serve.PprofToken 读取 pprof 访问令牌配置
func MainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // 记录请求开始时间

		// 获取或生成 requestID
		r, requestID := chain.Request(r)

		// 使用 `defer` 捕获 `panic`
		defer func() {
			if err := recover(); err != nil {
				// 记录 `panic` 错误日志
				slog.ErrorContext(r.Context(), "HTTP request handler panic error",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("start", start.Format(time.RFC3339Nano)),
					slog.Float64("elapsed", time.Since(start).Seconds()),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())), // 记录详细的堆栈信息
				)

				// 返回 自定义 panic error 服务器内部错误响应
				ResponseWriterPanicError(w)
			}

			// 记录请求完成日志
			slog.InfoContext(r.Context(), "HTTP request handler completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("start", start.Format(time.RFC3339Nano)),
				slog.Float64("elapsed", time.Since(start).Seconds()),
			)
		}()

		// set 具体的 X-Request-Id
		w.Header().Set(chain.XRquestID, requestID)

		// Step 1: Header 特殊处理逻辑, 主要为了放开跨域问题

		// 默认本服务处理の类型
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		// 防止 MIME 类型嗅探（MIME-Type Sniffing）
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 默认全添加 CORS 头, 默认允许跨域访问
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin) // 允许所有来源
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Request-Id, pprofbearer")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // 允许跨域携带 Cookie

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Step 2: pprof 路由的额外安全验证, 空 token 表示禁用认证
		if ispprofpath(r.URL.Path) && configs.G != nil {
			if pprofBearer := configs.G.Serve.PprofBearer; pprofBearer != "" && !verifypprofbearer(r, pprofBearer) {
				slog.WarnContext(r.Context(), "pprof access denied",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remoteAddr", r.RemoteAddr),
				)
				ResponseWriterPprofUnauthorized(w)
				return
			}
		}

		// Step 3: serve http call logic 之前额外操作

		next.ServeHTTP(w, r)

		// Step 4: serve http call logic 之后额外操作
	})
}
