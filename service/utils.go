package service

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang/protobuf/proto"
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