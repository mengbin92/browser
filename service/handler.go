package service

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/mengbin92/browser/utils"
	"google.golang.org/protobuf/runtime/protoiface"
)

func upload(ctx *gin.Context) {
	// new session
	session := sessions.Default(ctx)
	pbFile := uuid.NewString()
	session.Set("filename", pbFile)
	session.Save()

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		srvLogger.Errorf("FormFile error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("FormFile error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	defer file.Close()

	out, err := os.Create(utils.Fullname(pbFile))
	if err != nil {
		srvLogger.Errorf("create file: %s.pb error: %s", pbFile, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("create file: %s.pb error: %s", pbFile, err.Error()), "code": http.StatusInternalServerError})
		return
	}
	io.Copy(out, file)

	ctx.JSON(http.StatusCreated, gin.H{"msg": "upload file success", "code": http.StatusCreated})

	// cb := &common.Block{}
	// err = proto.Unmarshal(buf.Bytes(), cb)
	// if err != nil {
	// 	srvLogger.Errorf("Parse block error: %s", err.Error())
	// 	ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block error: %s", err.Error()), "code": http.StatusInternalServerError})
	// 	return
	// }

	// err = protolator.DeepMarshalJSON(buf, cb)
	// if err != nil {
	// 	srvLogger.Errorf("marshaler protobuf to json error: %s", err.Error())
	// 	ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("marshaler protobuf to json error: %s", err.Error()), "code": http.StatusInternalServerError})
	// 	return
	// }

	// ctx.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": buf.String()})
}

func sayHi(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "error" {
		// 返回kratos error
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "bad request"})
	} else {
		ctx.JSON(http.StatusOK, map[string]string{"welcome": name})
	}
}

func parse(ctx *gin.Context) {

	// get filename from session
	session := sessions.Default(ctx)
	pbFile := session.Get("filename")
	if pbFile == nil {
		srvLogger.Error("no filename in session")
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "no filename in session", "code": http.StatusBadRequest})
		return
	}

	msgType := ctx.Param("msgType")
	in, err := os.ReadFile(utils.Fullname(pbFile.(string)))
	if err != nil {
		srvLogger.Errorf("read file: %s.pb error: %s", pbFile, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("read file: %s error: %s", pbFile, err.Error()), "code": http.StatusBadRequest})
		return
	}

	var resp protoiface.MessageV1
	cb := &common.Block{}
	if err := proto.Unmarshal(in, cb); err != nil {
		srvLogger.Errorf("Parse block error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	switch msgType {
	case "block":
		resp = cb
	case "header":
		resp = cb.Header
	case "metadata":
		resp = cb.Metadata
	case "data":
		resp = cb.Data
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": fmt.Sprintf("unknow msgType: %s", msgType)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": resp})
}
