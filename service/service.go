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
	server    *Server
	pbcache   *pbCache
)

func init() {
	pbcache = &pbCache{
		cache: sync.Map{},
		time:  time.NewTicker(300 * time.Second),
	}

	go pbcache.checkExpiredTokenTimer()
}

type Server struct {
	srv *http.Server
}

func (s *Server) Routers() []*Router {
	return []*Router{
		{
			Path:     "/login",
			Method:   http.MethodPost,
			AuthType: config.Server_NOAUTH,
			Handler:  login,
		},
		{
			Path:     "/refresh",
			Method:   http.MethodGet,
			AuthType: config.Server_TOKENAUTH,
			Handler:  refresh,
		},
		{
			Path:     "/register",
			Method:   http.MethodPost,
			AuthType: config.Server_NOAUTH,
			Handler:  register,
		},
		{
			Path:     "/hi/:name",
			Method:   http.MethodGet,
			AuthType: config.Server_NOAUTH,
			Handler:  sayHi,
		},
		{
			Path:    "/block/upload",
			Method:  http.MethodPost,
			Handler: upload,
		},
		{
			Path:    "/block/parse/:msgType",
			Method:  http.MethodGet,
			Handler: parse,
		},
		{
			Path:    "/block/update/:channel",
			Method:  http.MethodPost,
			Handler: updateConfig,
		},
	}
}

func NewServer(conf *config.Server, logger *zap.SugaredLogger) {
	srvLogger = logger
	server = &Server{}

	engine := gin.Default()

	// using cookie store session
	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("pbFile", store))

	// engine.POST("/login", login)
	// engine.GET("/hi/:name", sayHi)
	// engine.POST("/block/upload", upload)
	// engine.GET("/block/parse/:msgType", parse)
	// engine.POST("/block/update/:channel", updateConfig)

	for _, router := range server.Routers() {
		var handlers []gin.HandlerFunc
		if router.AuthType == 0 {
			router.AuthType = conf.AuthType
		}
		switch router.AuthType {
		case config.Server_BASICAUTH:
			handlers = append(handlers, basicAuth)
		case config.Server_TOKENAUTH:
			handlers = append(handlers, tokenAuth)
		default:
			handlers = append(handlers, noAuth)
		}
		handlers = append(handlers, router.Handler)
		engine.Handle(router.Method, router.Path, handlers...)
	}

	server.srv = &http.Server{
		Addr:         conf.Http.Addr,
		Handler:      engine,
		ReadTimeout:  conf.Http.Timeout.AsDuration(),
		WriteTimeout: conf.Http.Timeout.AsDuration(),
	}
}

func Run() error {
	return server.srv.ListenAndServe()
}

func Shutdown() {
	srvLogger.Info("shutdown service")
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	server.srv.Shutdown(ctx)
}
