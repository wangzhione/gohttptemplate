// Package register provides initialization routines for the application environment.
package register

import (
	"context"
	"log/slog"
	"runtime"

	_ "net/http/pprof"

	"github.com/wangzhione/gohttptemplate/configs"
	"github.com/wangzhione/sbp/chain"
	"github.com/wangzhione/sbp/system"
)

// Init 启动之前的环境初始化 :) 必须是 once
func Init(ctx context.Context, path string) (err error) {
	// init config
	if err = configs.Init(ctx, path); err != nil {
		return
	}

	// slog 默认配置初始化
	switch configs.G.Log.Level {
	case "DEBUG":
		chain.EnableLevel = slog.LevelDebug
	case "INFO":
		chain.EnableLevel = slog.LevelInfo
	case "WARN":
		chain.EnableLevel = slog.LevelWarn
	case "ERROR":
		chain.EnableLevel = slog.LevelError
	}
	if err = chain.InitSLogRotatingFile(); err != nil {
		// 如果 文件 日志有问题, 需要打印相关信息
		slog.ErrorContext(ctx, "chain.InitSlogRotatingFile error", "error", err) // 退化成控制台输出

		chain.InitSLog() // 默认尝试退化成控制台输出
	}

	// 主动根据配置设置 GOMAXPROCS P 的数量, 用于特殊情况下限制服务机器 CPU 资源的使用
	if configs.G.Serve.GOMAXPROCS > 0 {
		runtime.GOMAXPROCS(configs.G.Serve.GOMAXPROCS)
	}

	slog.InfoContext(ctx, "main init start ...",
		slog.Time("SystemBeginTime", beginTime),
		slog.Int("cpunumber", runtime.NumCPU()),
		slog.Int("pnumber", runtime.GOMAXPROCS(0)),
		slog.String("path", path),
		slog.String("GOOS", runtime.GOOS),
		slog.String("BuildVersion", system.BuildVersion),
		slog.String("GitVersion", system.GitVersion),
		slog.String("GitCommitTime", system.GitCommitTime),
	)

	// 后续 init 操作, 放在 initlogic 里面
	if err = initlogic(ctx); err != nil {
		slog.ErrorContext(ctx, "initlogic error", "error", err)
		panic(err)
	}

	return
}
