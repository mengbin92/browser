package service

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/mengbin92/browser/utils"
	"github.com/pkg/errors"
)

type pbCache struct {
	cache sync.Map
	time  *time.Ticker
}

func (p *pbCache) checkExpiredTokenTimer() {
	for {
		<-p.time.C
		p.cache.Range(func(key, value interface{}) bool {
			pf := value.(*pbFile)
			if pf.isExpired() {
				p.cache.Delete(key)
				os.Remove(utils.Fullname(pf.Name))
			}
			return true
		})
	}
}

type pbFile struct {
	Name    string
	Expired int64
}

func (p *pbFile) isExpired() bool {
	return p.Expired < time.Now().Unix()
}

func (p *pbFile) renewal() {
	p.Expired = time.Now().Add(300 * time.Second).Unix()
}

func (p *pbFile) Marshal() ([]byte, error) {
	return sonic.Marshal(p)
}

func (p *pbFile) Unmarshal(data []byte) error {
	return sonic.Unmarshal(data, p)
}

type Creator struct {
	Mspid string
	Cert  *x509.Certificate
}

func getIdentity(serilizedIdentity []byte) (*Creator, error) {
	var err error

	sid := &msp.SerializedIdentity{}
	err = proto.Unmarshal(serilizedIdentity, sid)
	if err != nil {
		return nil, err
	}

	cert, err := decodeX509Pem(sid.IdBytes)
	if err != nil {
		return nil, errors.Wrap(err, "decodeX509Pem error")
	}

	return &Creator{
		Mspid: sid.Mspid,
		Cert:  cert,
	}, nil

}

func decodeX509Pem(certPem []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPem)
	if block == nil {
		return nil, errors.New("bad cert")
	}

	return x509.ParseCertificate(block.Bytes)
}

type Endorser struct {
	MSP  string `json:"msp"`
	Name string `json:"name"`
}

func marshalProtoMessage(pb proto.Message) ([]byte, error) {
	return proto.Marshal(pb)
}

func loadEnvelop(name string) (*common.Envelope, error) {
	in, err := os.ReadFile(utils.Fullname(name))
	if err != nil {
		srvLogger.Errorf("read file: %s.pb error: %s", name, err.Error())
		return nil, errors.Wrapf(err, "read file: %s", name)
	}

	cb := &common.Block{}
	if err := proto.Unmarshal(in, cb); err != nil {
		srvLogger.Errorf("Parse block error: %s", err.Error())
		return nil, errors.Wrap(err, "Parse block error")
	}
	env := &common.Envelope{}
	if err := proto.Unmarshal(cb.Data.Data[0], env); err != nil {
		srvLogger.Errorf("Parse block data error: %s", err.Error())
		return nil, errors.Wrap(err, "Parse block data error")
	}
	return env, nil
}

func loadSession(ctx *gin.Context) (string, error) {
	// get filename from session
	session := sessions.Default(ctx)
	buf := session.Get("filename")
	if buf == nil {
		srvLogger.Error("no filename in session")
		return "", errors.New("no filename in session")
	}

	// 更新pbFile过期时间
	pf := &pbFile{}
	pf.Unmarshal([]byte(buf.(string)))
	pf.renewal()

	data, _ := pf.Marshal()
	session.Set("filename", string(data))
	session.Save()
	return pf.Name, nil
}
