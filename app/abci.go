package app

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"

	"github.com/saguywalker/sitcomchain/code"
	protoTm "github.com/saguywalker/sitcomchain/proto/tendermint"
)

// Info return current information of blockchain
func (a *SitcomApplication) Info(req types.RequestInfo) (res types.ResponseInfo) {
	res.Version = a.Version
	res.AppVersion = a.AppProtocolVersion
	res.LastBlockHeight = a.state.Height
	res.LastBlockAppHash = a.state.AppHash
	return res
}

// DeliverTx update new data
func (a *SitcomApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	a.logger.Infof("In deliverTx: %s\n", string(req.Tx))

	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	payload := txObj.Payload
	var err error
	switch payload.Method {
	case "SetValidator":
		res = a.setValidator(string(payload.Params))
		a.state.Size++
	case "GiveBadge":
		res, err = a.giveBadge(payload.Params)
		if err != nil {
			return
		}
	case "ApproveActivity":
		res, err = a.approveActivity(payload.Params)
		if err != nil {
			return
		}
	default:
		res.Log = fmt.Sprintf("unknown method %s", payload.Method)
		res.Code = code.CodeTypeInvalidMethod
	}

	return res
}

// CheckTx validate data format before putting in mempool
func (a *SitcomApplication) CheckTx(req types.RequestCheckTx) (res types.ResponseCheckTx) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	a.logger.Infof("In checkTx: %s\n", string(req.Tx))

	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		a.logger.Errorln("Error in unmarshal txObj")
		res.Code = code.CodeTypeUnmarshalError
		res.Log = err.Error()
		return
	}

	payload := txObj.Payload
	pubKey := txObj.PublicKey
	signature := txObj.Signature

	if payload.Method == "" {
		res.Code = code.CodeTypeEmptyMethod
		res.Log = "method cannot be emptry"
		return
	}

	if _, exists := methodList[payload.Method]; !exists {
		res.Code = code.CodeTypeInvalidMethod
		res.Log = fmt.Sprintf("unknown method: %s", payload.Method)
		return
	}

	a.logger.Infof("pubkey: 0x%x\nparams: %s\nsignature: 0x%x\n", pubKey, payload.Params, signature)

	if !ed25519.Verify(pubKey, payload.Params, signature) {
		a.logger.Errorf("Failed in signature verification\n")
		res.Code = code.CodeTypeUnauthorized
		res.Log = "failed in signature verification"
		return
	}

	res.Code = code.CodeTypeOK
	return
}

// Commit commit a current transaction batch
func (a *SitcomApplication) Commit() (res types.ResponseCommit) {
	AppHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(AppHash, a.state.Size)

	a.state.Height++
	a.state.AppHash = AppHash
	a.state.SaveState()

	return
}

// Query return data from blockchain
func (a *SitcomApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	a.logger.Infof("In query: %s\n", string(req.Data))

	if len(req.Data) == 0 {
		itr := a.state.db.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			a.logger.Printf("k: %s, v: %s\n", itr.Key(), itr.Value())
		}
		return
	}

	// For query
	res.Key = req.Data
	parts := bytes.Split(res.Key, []byte("="))
	if len(parts) == 2 {
		result := make([]byte, 0)
		itr := a.state.db.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			key := itr.Key()
			if bytes.Contains(key, parts[0]) && bytes.Contains(key, parts[1]) {
				result = append(result, []byte("|")...)
				result = append(result, key...)
			}
		}

		if len(result) == 0 {
			res.Log = "does not exists"
			return
		}

		res.Log = "exists"
		res.Value = result[1:]
		return

	}

	// For verify
	value := a.state.db.Get(req.Data)
	if value != nil {
		res.Log = "exists"
		res.Value = value
		return
	}

	res.Log = "does not exist"
	return
}

// InitChain is used for initialize a blockchain
func (a *SitcomApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := a.updateValidator(v)
		if r.IsErr() {
			a.logger.Errorf("Error updating validators: %v", r)
		}
	}
	return types.ResponseInitChain{}
}

// BeginBlock create new transaction batch
func (a *SitcomApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	a.logger.Infof("BeginBlock: %d, ChainID: %s", req.Header.Height, req.Header.ChainID)
	a.state.Height = req.Header.Height
	a.CurrentChain = req.Header.ChainID
	a.valUpdates = make(map[string]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// EndBlock is called when ending block
func (a *SitcomApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	a.logger.Infof("EndBlock: %d", req.Height)
	valUpdates := make([]types.ValidatorUpdate, 0)
	for _, v := range a.valUpdates {
		valUpdates = append(valUpdates, v)
	}
	return types.ResponseEndBlock{ValidatorUpdates: valUpdates}
}
