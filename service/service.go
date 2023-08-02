package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/mengbin92/browser/conf"
	"go.uber.org/zap"
)

var (
	srvLogger *zap.SugaredLogger
	server    *http.Server
)

func NewServer(conf *conf.Server, logger *zap.SugaredLogger) {
	srvLogger = logger

	engine := gin.Default()

	// using cookie store session
	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("pbFile", store))

	engine.GET("/hi/:name", sayHi)
	engine.POST("/block/upload", upload)
	engine.GET("/block/parse/:msgType", parse)

	server = &http.Server{
		Addr:    conf.Http.Addr,
		Handler: engine,
	}
}

func Run() error {
	return server.ListenAndServe()
}

func Shutdown() {
	srvLogger.Info("shutdown service")
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	server.Shutdown(ctx)
}
