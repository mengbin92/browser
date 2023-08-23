package service

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"
)

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
