package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mengbin92/browser/data"
	"github.com/mengbin92/browser/utils"
)

type Login struct {
	Name     string `form:"name"`
	Password string	`form:"password"`
}

func login(ctx *gin.Context) {
	login := &Login{}
	if err := ctx.Bind(login); err != nil {
		srvLogger.Errorf("Binding Login info error: %s\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := data.Login(login.Name, login.Password)
	if err != nil {
		srvLogger.Errorf("login user: %s error: %s", login.Name, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": resp})
}

func register(ctx *gin.Context) {
	login := &Login{}
	if err := ctx.Bind(login); err != nil {
		srvLogger.Errorf("Binding Login info error: %s\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := data.RegisterUser(login.Name, login.Password)
	if err != nil {
		srvLogger.Errorf("register user: %s error: %s", login.Name, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": resp})
}

func noAuth(ctx *gin.Context) {
	ctx.Next()
}

func basicAuth(ctx *gin.Context) {
	name, pwd, ok := ctx.Request.BasicAuth()
	if !ok {
		srvLogger.Error("basic auth failed")
		ctx.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "basic auth failed"})
		ctx.Abort()
		return
	}
	user, err := data.GetUserByName(name)
	if err != nil {
		srvLogger.Errorf("GetUserByName error: %s", err.Error())
		ctx.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": err.Error()})
		ctx.Abort()
		return
	}
	if utils.CalcPassword(pwd, user.Salt) != user.Password {
		srvLogger.Error("user name or password is incorrect")
		ctx.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "user name or password is incorrect"})
		ctx.Abort()
		return
	}
	ctx.Next()
}

func tokenAuth(ctx *gin.Context) {
	if err := data.ParseJWT(strings.Split(ctx.Request.Header.Get("Authorization"), " ")[1]); err != nil {
		srvLogger.Errorf("tokenAuth error: %s", err.Error())
		ctx.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "token auth failed"})
		ctx.Abort()
		return
	}
	ctx.Next()
}
