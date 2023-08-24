package server

import (
	"context"
	v1 "mengbin92/browser/api/browser/v1"
	"mengbin92/browser/internal/conf"
	"mengbin92/browser/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, browser *service.BrowserService, block *service.BlockService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return []byte(c.Auth.JwtSecret), nil
				}),
			).Match(whiteMatcher()).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterBrowserHTTPServer(srv, browser)
	v1.RegisterBlockHTTPServer(srv, block)
	return srv
}

func whiteMatcher() selector.MatchFunc {
	whitelist := make(map[string]struct{})
	whitelist["/api.browser.v1.Browser/GetToken"] = struct{}{}
	whitelist["/api.browser.v1.Browser/Regisger"] = struct{}{}
	return func(ctx context.Context, path string) bool {
		if _, ok := whitelist[path]; ok {
			return false
		}
		return true
	}
}
