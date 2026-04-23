package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/wangzhione/gohttptemplate/configs"
	"github.com/wangzhione/gohttptemplate/handler/middleware"
	"github.com/wangzhione/gohttptemplate/register"
	"github.com/wangzhione/sbp/chain"
	"github.com/wangzhione/sbp/https"
)

var (
	// fpath 默认配置文件地址
	fpath = flag.String("f", "resource/etc/prod.toml", "The config file")

	// addr 默认监听地址, 用于 -a 输入服务监听地址
	addrs = flag.String("a", "", "The address to listen on or (0.0.0.0:8089)")
)

func main() {
	flag.Parse() // flag 参数初始化

	ctx := chain.Context()
	defer https.End(ctx)

	// init 如果失败, 程序会直接退出
	err := register.Init(ctx, *fpath)
	if err != nil {
		os.Exit(-1)
	}

	serverAddr := *addrs
	if serverAddr == "" {
		serverAddr = fmt.Sprintf("0.0.0.0:%d", configs.G.Serve.Port) // 0.0.0.0 默认 ipv4 绑定本机地址
	}

	https.ServeLoop(
		ctx,
		serverAddr,
		middleware.MainMiddleware(http.DefaultServeMux),
		configs.G.Serve.StopTime,
	)
}
