package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"

	"github.com/saguywalker/sitcomchain/code"
	protoTm "github.com/saguywalker/sitcomchain/proto/tendermint"
)

// Info return current information of blockchain
func (app *SitcomApplication) Info(req types.RequestInfo) (res types.ResponseInfo) {
	res.Version = app.Version
	res.AppVersion = app.AppProtocolVersion
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	return res
}

// DeliverTx update new data
func (app *SitcomApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	app.logger.Infof("In deliverTx: %s\n", string(req.Tx))

	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	payload := txObj.Payload

	switch payload.Method {
	case "SetValidator":
		res = app.setValidator(string(payload.Params))
		app.state.Size++
	case "GiveBadge":
		var sorted map[string]interface{}
		if err := json.Unmarshal(payload.Params, &sorted); err != nil {
			res.Code = code.CodeTypeUnmarshalError
			res.Log = "error when unmarshal params"
			break
		}

		delete(sorted, "giver")

		badgeKey, err := json.Marshal(sorted)
		if err != nil {
			res.Code = code.CodeTypeEncodingError
			res.Log = "error when marshal badgeKey"
			break
		}

		app.logger.Infof("k: %s, v: %s\n", badgeKey, payload.Params)
		app.state.db.Set(badgeKey, payload.Params)
		app.state.Size++
		res.Code = code.CodeTypeOK
		res.Log = "success"
	default:
		res.Log = fmt.Sprintf("unknown method %s", payload.Method)
		res.Code = code.CodeTypeInvalidMethod
	}

	return res
}

// CheckTx validate data format before putting in mempool
func (app *SitcomApplication) CheckTx(req types.RequestCheckTx) (res types.ResponseCheckTx) {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	app.logger.Infof("In checkTx: %s\n", string(req.Tx))

	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		app.logger.Errorln("Error in unmarshal txObj")
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

	app.logger.Infof("pubkey: 0x%x\nparams: %s\nsignature: 0x%x\n", pubKey, payload.Params, signature)

	if !ed25519.Verify(pubKey, payload.Params, signature) {
		app.logger.Errorf("Failed in signature verification\n")
		res.Code = code.CodeTypeUnauthorized
		res.Log = "failed in signature verification"
		return
	}

	res.Code = code.CodeTypeOK
	return
}

// Commit commit a current transaction batch
func (app *SitcomApplication) Commit() (res types.ResponseCommit) {
	appHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(appHash, app.state.Size)

	app.state.Height++
	app.state.AppHash = appHash
	app.state.SaveState()

	return
}

// Query return data from blockchain
func (app *SitcomApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	defer func() {
		if r := recover(); r != nil {
			app.logger.Errorln(r)
			res.Code = code.CodeTypeUnknownError
		}
	}()

	app.logger.Infof("In query: %s\n", string(req.Data))

	if len(req.Data) == 0 {
		itr := app.state.db.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			app.logger.Printf("k: %s, v: %s\n", itr.Key(), itr.Value())
		}
		return
	}

	// For query
	res.Key = req.Data
	parts := bytes.Split(res.Key, []byte("="))
	if len(parts) == 2 {
		result := make([]byte, 0)
		itr := app.state.db.Iterator(nil, nil)
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
	value := app.state.db.Get(req.Data)
	if value != nil {
		res.Log = "exists"
		res.Value = value
		return
	}

	res.Log = "does not exist"
	return
}

// InitChain is used for initialize a blockchain
func (app *SitcomApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Errorf("Error updating validators: %v", r)
		}
	}
	return types.ResponseInitChain{}
}

// BeginBlock create new transaction batch
func (app *SitcomApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Infof("BeginBlock: %d, ChainID: %s", req.Header.Height, req.Header.ChainID)
	app.state.Height = req.Header.Height
	app.CurrentChain = req.Header.ChainID
	app.valUpdates = make(map[string]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// EndBlock is called when ending block
func (app *SitcomApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logger.Infof("EndBlock: %d", req.Height)
	valUpdates := make([]types.ValidatorUpdate, 0)
	for _, v := range app.valUpdates {
		valUpdates = append(valUpdates, v)
	}
	return types.ResponseEndBlock{ValidatorUpdates: valUpdates}
}
