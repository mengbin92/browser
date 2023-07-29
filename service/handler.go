package service

import (
	"github.com/gin-gonic/gin"
	kgin "github.com/go-kratos/gin"
	"github.com/go-kratos/kratos/v2/errors"
)

func uploadFile(ctx *gin.Context) {

}

func getConfig(ctx *gin.Context) {

}

func getBlock(ctx *gin.Context) {

}

func sayHi(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "error" {
		// 返回kratos error
		kgin.Error(ctx, errors.Unauthorized("auth_error", "no authentication"))
	} else {
		ctx.JSON(200, map[string]string{"welcome": name})
	}
}
