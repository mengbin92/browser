package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-protos-go/common"
)

func main() {
	// loadConfigToJSON()
	loadBlockToJSON()
}

func loadConfigToJSON() {
	data, err := os.ReadFile("mychannel_config.block")
	if err != nil {
		panic(err)
	}
	cb := &common.Block{}
	err = proto.Unmarshal(data, cb)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	if err := protolator.DeepMarshalJSON(buf, cb);err != nil{
		panic(err)
	}
	fmt.Println(buf.String())
}

func loadBlockToJSON(){
	data, err := os.ReadFile("mychannel_newest.block")
	if err != nil {
		panic(err)
	}
	cb := &common.Block{}
	err = proto.Unmarshal(data, cb)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	if err := protolator.DeepMarshalJSON(buf, cb);err != nil{
		panic(err)
	}
	fmt.Println(buf.String())
}