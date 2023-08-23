package service

import (
	"bytes"
	"context"
	"io"
	"os"

	pb "mengbin92/browser/api/browser/v1"
	"mengbin92/browser/internal/utils"

	"github.com/bytedance/sonic"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/runtime/protoiface"
)

type BlockService struct {
	pb.UnimplementedBlockServer
}

func NewBlockService() *BlockService {
	return &BlockService{}
}

func (s *BlockService) UpChaincode(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	_, err := os.Stat("./pb")
	if os.IsNotExist(err) {
		os.Mkdir("./pb", 0750)
	}

	out, err := os.Create(utils.Fullname(req.Name))
	if err != nil {
		log.Errorf("create file: %s error: %s", utils.Fullname(req.Name), err.Error())
		return &pb.UploadResponse{
			Result: false,
		}, errors.Wrapf(err, "create file: %s error", req.Name)
	}
	out.Write(req.Content)

	return &pb.UploadResponse{
		Result: true,
		Name:   req.Name,
	}, nil
}
func (s *BlockService) ParseBlock(ctx context.Context, req *pb.ParseRequest) (*pb.ParseResponse, error) {
	file, err := os.Open(utils.Fullname(req.Name))
	if err != nil {
		log.Errorf("open file: %s error: %s", utils.Fullname(req.Name), err.Error())
		return nil, errors.Wrap(err, "load block data error")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("read file: %s error: %s", utils.Fullname(req.Name), err.Error())
		return nil, errors.Wrap(err, "read block data error")
	}

	var resp protoiface.MessageV1
	buf := &bytes.Buffer{}

	blk := &common.Block{}
	if err := proto.Unmarshal(data, blk); err != nil {
		log.Errorf("load block struct from data error: %s", err.Error())
		return nil, errors.Wrap(err, "load block struct from data error")
	}
	env := &common.Envelope{}
	if err := proto.Unmarshal(blk.Data.Data[0], env); err != nil {
		log.Errorf("Parse block data error: %s", err.Error())
		return nil, errors.Wrap(err, "Parse block data error")
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(env.Payload, payload); err != nil {
		log.Errorf("Parse block Payload error: %s", err.Error())
		return nil, errors.Wrap(err, "Parse block Payload error")
	}
	switch req.Operation {
	case pb.ParseRequest_ACTIONS:
	case pb.ParseRequest_TRANSACTION:
		resp, err = utils.GetTransaction(payload.Data)
		if err != nil {
			log.Errorf("Parse block Transaction error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block Transaction error")
		}
	case pb.ParseRequest_HEADER:
		resp = blk.Header
	case pb.ParseRequest_METADATA:
		resp = blk.Metadata
	case pb.ParseRequest_DATA:
		resp = blk.Data
	case pb.ParseRequest_CONFIG:
		resp, err = utils.GetConfigEnvelope(payload.Data)
		if err != nil {
			log.Errorf("Parse block ConfigEnvelope error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block ConfigEnvelope error")
		}
	case pb.ParseRequest_CHAINCODE:
		resp, err = utils.GetChaincodeHeaderExtension(payload.Header)
		if err != nil {
			log.Errorf("Parse block ChaincodeHeaderExtension error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block ChaincodeHeaderExtension error")
		}
	case pb.ParseRequest_INPUT:
		chaincodeProposalPayload, _, _, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			log.Errorf("Parse block ChaincodeInvocationSpec error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block ChaincodeInvocationSpec error")
		}
		resp, err = utils.GetChaincodeInvocationSpec(chaincodeProposalPayload.Input)
		if err != nil {
			log.Errorf("Parse block ChaincodeInvocationSpec input error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block ChaincodeInvocationSpec input error")
		}
	case pb.ParseRequest_RWSET:
		_, _, chaincodeAction, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			log.Errorf("Parse block ChaincodeAction error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block ChaincodeAction error")
		}
		resp, err = utils.GetRWSet(chaincodeAction)
		if err != nil {
			log.Errorf("Parse block TxReadWriteSet error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block TxReadWriteSet error")
		}
	case pb.ParseRequest_CHANNEL:
		resp, err = utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			log.Errorf("Parse block UnmarshalChannelHeader error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block UnmarshalChannelHeader error")
		}
	case pb.ParseRequest_ENDORSEMENTS:
		_, endorsements, _, err := utils.ParseChaincodeEnvelope(env)
		if err != nil {
			log.Errorf("Parse block Endorsement error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block Endorsement error")
		}
		endorsers := &pb.Endorsers{
			Endorsers: make([]*pb.Endorser, len(endorsements)),
		}

		for index, e := range endorsements {
			identity, err := getIdentity(e.Endorser)
			if err != nil {
				if err != nil {
					log.Errorf("Parse Identity error: %s", err.Error())
					return nil, errors.Wrap(err, "Parse Identity error")
				}
			}
			userName := ""
			if identity.Cert != nil {
				userName = identity.Cert.Subject.CommonName
			}
			endorsers.Endorsers[index] = &pb.Endorser{MSP: identity.Mspid, Name: userName}
		}
		resp = endorsers
	case pb.ParseRequest_CREATOR:
		shdr, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
		if err != nil {
			log.Errorf("Parse block SignatureHeader error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block SignatureHeader error")
		}
		creator, err := getIdentity(shdr.Creator)
		if err != nil {
			log.Errorf("Parse block Creator error: %s", err.Error())
			return nil, errors.Wrap(err, "Parse block Creator error")
		}
		data, err := sonic.Marshal(creator)
		if err != nil {
			log.Errorf("Marshal Creator error: %s", err.Error())
			return nil, errors.Wrap(err, "Marshal Creator error")
		}
		io.WriteString(buf, string(data))
	default:
		resp = blk
	}
	if resp != nil {
		if err := protolator.DeepMarshalJSON(buf, resp); err != nil {
			log.Errorf("DeepMarshalJSON data error: %s", err.Error())
			return nil, errors.Wrap(err, "DeepMarshalJSON data error")
		}
	}

	return &pb.ParseResponse{}, nil
}
