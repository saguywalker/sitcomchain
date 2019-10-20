package app

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"

	"github.com/saguywalker/sitcomchain/code"
	protoTm "github.com/saguywalker/sitcomchain/proto/tendermint"
)

// Info return current information of blockchain
func (a *SitcomApplication) Info(req types.RequestInfo) (res types.ResponseInfo) {
	res.LastBlockHeight = a.state.Height
	res.LastBlockAppHash = a.state.AppHash
	return res
}

// DeliverTx update new data
func (a *SitcomApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {
	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	payload := txObj.Payload

	switch payload.Method {
	case "SetValidator":
		res = a.setValidator(payload.Params)
	case "GiveBadge":
		a.state.currentBatch.Set([]byte(payload.Params), []byte(payload.Params))
		res.Code = code.CodeTypeOK
		res.Log = "success"
	default:
		res.Log = fmt.Sprintf("unknown method %s", payload.Method)
		res.Code = code.CodeTypeInvalidMethod
	}

	return res
}

// CheckTx validate data format before putting in mempool
func (a *SitcomApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		return types.ResponseCheckTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	payload := txObj.Payload
	pubKey := txObj.PublicKey
	signature := txObj.Signature

	if payload.Method == "" {
		return types.ResponseCheckTx{
			Code: code.CodeTypeEmptyMethod,
			Log:  "method cannot be empty"}
	}

	if _, exists := methodList[payload.Method]; !exists {
		return types.ResponseCheckTx{
			Code: code.CodeTypeInvalidMethod,
			Log:  fmt.Sprintf("unknown for method %s", payload.Method)}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return types.ResponseCheckTx{
			Code: code.CodeTypeEncodingError,
			Log:  "error with payload unmarshal"}
	}

	if !ed25519.Verify(pubKey, payloadBytes, signature) {
		return types.ResponseCheckTx{
			Code: code.CodeTypeUnauthorized,
			Log:  "failed in signature verification",
		}
	}

	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

// Commit commit a current transaction batch
func (a *SitcomApplication) Commit() types.ResponseCommit {
	a.state.currentBatch.Commit()
	return types.ResponseCommit{Data: []byte{}}
}

// Query return data from blockchain
func (a *SitcomApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	res.Key = req.Data
	err := a.state.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			res.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				res.Log = "exists"
				res.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitChain is used for initialize a blockchain
func (a *SitcomApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

// BeginBlock create new transaction batch
func (a *SitcomApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	a.state.currentBatch = a.state.db.NewTransaction(true)
	return types.ResponseBeginBlock{}
}

// EndBlock
func (a *SitcomApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}
