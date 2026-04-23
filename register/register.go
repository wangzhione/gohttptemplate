package register

import (
	"context"
	"log/slog"
	"time"

	// init logic server register
	_ "github.com/wangzhione/gohttptemplate/internal/logic"
)

// beginTime 记录当前进程启动初始化时间，避免依赖 sbp 内部变量位置变更。
var beginTime = time.Now()

func initlogic(ctx context.Context) (err error) {
	defer func() {
		end := time.Now()

		slog.InfoContext(ctx, "logic init end",
			slog.Any("reason", err),
			slog.Float64("elapsed_seconds", end.Sub(beginTime).Seconds()),
		)
	}()

	// do something register logic init here 👇

	// mysql init
	// if err = mysqllogic.Init(ctx, configs.G.MySQL.Main); err != nil {
	// 	return err
	// }

	return
}
