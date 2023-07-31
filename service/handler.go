package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-protos-go/common"
)

func parseBlock(ctx *gin.Context) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		srvLogger.Errorf("FormFile error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("FormFile error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	defer file.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, file)

	cb := &common.Block{}
	err = proto.Unmarshal(buf.Bytes(), cb)
	if err != nil {
		srvLogger.Errorf("Parse block error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}

	err = protolator.DeepMarshalJSON(buf, cb)
	if err != nil {
		srvLogger.Errorf("marshaler protobuf to json error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("marshaler protobuf to json error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"code": http.StatusCreated, "msg": buf.String()})
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
