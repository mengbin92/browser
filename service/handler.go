package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
	uid := uuid.NewString()
	pf := &pbFile{
		Name:    uid,
		Expired: time.Now().Add(300 * time.Second).Unix(),
	}
	buf, _ := pf.Marshal()
	session.Set("filename", string(buf))
	session.Save()

	pbcache.cache.Store(uid, pf)

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		srvLogger.Errorf("FormFile error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("FormFile error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	defer file.Close()

	out, err := os.Create(utils.Fullname(uid))
	if err != nil {
		srvLogger.Errorf("create file: %s.pb error: %s", uid, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("create file: %s.pb error: %s", uid, err.Error()), "code": http.StatusInternalServerError})
		return
	}
	io.Copy(out, file)

	ctx.JSON(http.StatusCreated, gin.H{"msg": "upload file success", "code": http.StatusCreated})
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
	buf := session.Get("filename")
	if buf == nil {
		srvLogger.Error("no filename in session")
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "no filename in session", "code": http.StatusBadRequest})
		return
	}

	// 更新pbFile过期时间
	pf := &pbFile{}
	pf.Unmarshal([]byte(buf.(string)))
	pf.renewal()

	data, _ := pf.Marshal()
	session.Set("filename", string(data))
	session.Save()

	msgType := ctx.Param("msgType")
	in, err := os.ReadFile(utils.Fullname(pf.Name))
	if err != nil {
		srvLogger.Errorf("read file: %s.pb error: %s", pf.Name, err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("read file: %s error: %s", pf.Name, err.Error()), "code": http.StatusBadRequest})
		return
	}

	var resp protoiface.MessageV1
	cb := &common.Block{}
	if err := proto.Unmarshal(in, cb); err != nil {
		srvLogger.Errorf("Parse block error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	env := &common.Envelope{}
	if err := proto.Unmarshal(cb.Data.Data[0], env); err != nil {
		srvLogger.Errorf("Parse block data error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block data error: %s", err.Error()), "code": http.StatusInternalServerError})
		return
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(env.Payload, payload); err != nil {
		srvLogger.Errorf("Parse block Payload error: %s", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block Payload error: %s", err.Error()), "code": http.StatusInternalServerError})
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
	case "config":
		resp, err = utils.GetConfigEnvelope(payload.Data)
		if err != nil {
			srvLogger.Errorf("Parse block ConfigEnvelope error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block ConfigEnvelope error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
	case "chaincode":
		resp, err = utils.GetChaincodeHeaderExtension(payload.Header)
		if err != nil {
			srvLogger.Errorf("Parse block ChaincodeHeaderExtension error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block ChaincodeHeaderExtension error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
	case "actions":
	case "transaction":
		resp, err = utils.GetTransaction(payload.Data)
		if err != nil {
			srvLogger.Errorf("Parse block Transaction error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block Transaction error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
	case "input":
		chaincodeProposalPayload, _, _, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			srvLogger.Errorf("Parse block ChaincodeInvocationSpec error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block ChaincodeInvocationSpec error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
		resp,err = utils.GetChaincodeInvocationSpec(chaincodeProposalPayload.Input)
		if err != nil {
			srvLogger.Errorf("Parse block ChaincodeInvocationSpec input error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block ChaincodeInvocationSpec input error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
	case "rwset":
		_,_,chaincodeAction, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			srvLogger.Errorf("Parse block ChaincodeAction error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block ChaincodeAction error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
		resp,err = utils.GetRWSet(chaincodeAction)
		if err != nil {
			srvLogger.Errorf("Parse block TxReadWriteSet error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block TxReadWriteSet error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
	case "endorsements":
		_, endorsements, _, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			srvLogger.Errorf("Parse block Endorsement error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block Endorsement error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
		endors := make([]*Endorser, len(endorsements))
		for _, e := range endorsements {
			identity, err := getIdentity(e.Endorser)
			if err != nil {
				if err != nil {
					srvLogger.Errorf("Parse Identity error: %s", err.Error())
					ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse Identity error: %s", err.Error()), "code": http.StatusInternalServerError})
					return
				}
			}
			userName := ""
			if identity.Cert != nil {
				userName = identity.Cert.Subject.CommonName
			}
			endors = append(endors, &Endorser{MSP: identity.Mspid, Name: userName})
		}
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": endors})
		return
	case "creator":
		shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
		if err != nil {
			srvLogger.Errorf("Parse block SignatureHeader error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block SignatureHeader error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
		creator, err := getIdentity(shdr.Creator)
		if err != nil {
			srvLogger.Errorf("Parse block Creator error: %s", err.Error())
			ctx.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Parse block Creator error: %s", err.Error()), "code": http.StatusInternalServerError})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": creator})
		return
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "msg": fmt.Sprintf("unknow msgType: %s", msgType)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": resp})
}
