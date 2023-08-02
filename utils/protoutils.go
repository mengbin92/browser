package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// GetPayload Get Payload from Envelope message
func GetPayload(e *common.Envelope) (*common.Payload, error) {
	payload := &common.Payload{}
	err := proto.Unmarshal(e.Payload, payload)
	return payload, errors.Wrap(err, "error unmarshaling Payload")
}

// UnmarshalChannelHeader returns a ChannelHeader from bytes
func UnmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	chdr := &common.ChannelHeader{}
	err := proto.Unmarshal(bytes, chdr)
	return chdr, errors.Wrap(err, "error unmarshaling ChannelHeader")
}

// GetSignatureHeader Get SignatureHeader from bytes
func GetSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	return UnmarshalSignatureHeader(bytes)
}

// UnmarshalSignatureHeader unmarshals bytes to a SignatureHeader.
func UnmarshalSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	if err := proto.Unmarshal(bytes, sh); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling SignatureHeader")
	}
	return sh, nil
}

// GetChaincodeHeaderExtension get chaincode header extension given header
func GetChaincodeHeaderExtension(hdr *common.Header) (*peer.ChaincodeHeaderExtension, error) {
	chdr, err := UnmarshalChannelHeader(hdr.ChannelHeader)
	if err != nil {
		return nil, err
	}

	chaincodeHdrExt := &peer.ChaincodeHeaderExtension{}
	err = proto.Unmarshal(chdr.Extension, chaincodeHdrExt)
	return chaincodeHdrExt, errors.Wrap(err, "error unmarshaling ChaincodeHeaderExtension")
}

// GetTransaction Get Transaction from bytes
func GetTransaction(txBytes []byte) (*peer.Transaction, error) {
	tx := &peer.Transaction{}
	err := proto.Unmarshal(txBytes, tx)
	return tx, errors.Wrap(err, "error unmarshaling Transaction")

}

// GetProposalResponsePayload gets the proposal response payload
func GetProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, errors.Wrap(err, "error unmarshaling ProposalResponsePayload")
}

// GetChaincodeActionPayload Get ChaincodeActionPayload from bytes
func GetChaincodeActionPayload(capBytes []byte) (*peer.ChaincodeActionPayload, error) {
	cap := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(capBytes, cap)
	return cap, errors.Wrap(err, "error unmarshaling ChaincodeActionPayload")
}

// GetChaincodeAction gets the ChaincodeAction given chaicnode action bytes
func GetChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, errors.Wrap(err, "error unmarshaling ChaincodeAction")
}

// GetPayloads gets the underlying payload objects in a TransactionAction
func GetPayloads(txActions *peer.TransactionAction) (*peer.ChaincodeActionPayload, *peer.ChaincodeAction, error) {
	// TODO: pass in the tx type (in what follows we're assuming the
	// type is ENDORSER_TRANSACTION)
	ccPayload, err := GetChaincodeActionPayload(txActions.Payload)
	if err != nil {
		return nil, nil, err
	}

	if ccPayload.Action == nil || ccPayload.Action.ProposalResponsePayload == nil {
		return nil, nil, errors.New("no payload in ChaincodeActionPayload")
	}
	pRespPayload, err := GetProposalResponsePayload(ccPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, nil, err
	}

	if pRespPayload.Extension == nil {
		return nil, nil, errors.New("response payload is missing extension")
	}

	respPayload, err := GetChaincodeAction(pRespPayload.Extension)
	if err != nil {
		return ccPayload, nil, err
	}
	return ccPayload, respPayload, nil
}

// GetChaincodeProposalPayload Get ChaincodeProposalPayload from bytes
func GetChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshaling ChaincodeProposalPayload")
}
