package service

import (
	"os"

	"github.com/gin-gonic/gin"
	kgin "github.com/go-kratos/gin"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/mengbin92/browser/conf"
)

var (
	app       *kratos.App
	srvLogger *log.Helper
)

func NewServer(conf *conf.Server, logger log.Logger) {
	srvLogger = log.NewHelper(logger)
	id, _ := os.Hostname()

	router := gin.Default()

	// 使用kratos中间件
	router.Use(kgin.Middlewares(recovery.Recovery(), customMiddleware))

	router.GET("/hi/:name", sayHi)

	httpSrv := http.NewServer(http.Address(conf.Http.Addr))
	httpSrv.HandlePrefix("/", router)

	app = kratos.New(
		kratos.Name("gin-blockchain-browser"),
		kratos.ID(id),
		kratos.Version("1.0.0."),
		kratos.Server(httpSrv),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
	)
}

func Run() error {
	return app.Run()
}

func Stop() error {
	return app.Stop()
}
