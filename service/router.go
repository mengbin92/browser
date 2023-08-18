package service

import (
	"github.com/gin-gonic/gin"
	"github.com/mengbin92/browser/conf"
)

type Router struct {
	Path     string
	Method   string
	AuthType conf.Server_AuthType
	Handler  gin.HandlerFunc
}
