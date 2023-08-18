package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	config "github.com/mengbin92/browser/conf"
	"go.uber.org/zap"
)

var (
	srvLogger *zap.SugaredLogger
	server    *http.Server
	pbcache   *pbCache
)

func init() {
	pbcache = &pbCache{
		cache: sync.Map{},
		time:  time.NewTicker(300 * time.Second),
	}

	go pbcache.checkExpiredTokenTimer()
}

func NewServer(conf *config.Server, logger *zap.SugaredLogger) {
	srvLogger = logger

	engine := gin.Default()

	// using cookie store session
	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("pbFile", store))

	engine.POST("/login", login)
	engine.GET("/hi/:name", sayHi)

	var handlers []gin.HandlerFunc

	switch conf.AuthType {
	case config.Server_BASICAUTH:
		handlers = append(handlers, basicAuth)
	case config.Server_TOKENAUTH:
		handlers = append(handlers, tokenAuth)
	default:
		handlers = append(handlers, noAuth)
	}

	engine.POST("/block/upload", []gin.HandlerFunc{handlers[0], upload}...)
	engine.GET("/block/parse/:msgType", []gin.HandlerFunc{handlers[0], parse}...)
	engine.POST("/block/update/:channel", []gin.HandlerFunc{handlers[0], updateConfig}...)

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
