package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"

	"github.com/saguywalker/sitcomchain/code"
	protoTm "github.com/saguywalker/sitcomchain/proto/tendermint"
)

// Info return current information of blockchain
func (app *SitcomApplication) Info(req types.RequestInfo) (res types.ResponseInfo) {
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	return res
}

// DeliverTx update new data
func (app *SitcomApplication) DeliverTx(req types.RequestDeliverTx) (res types.ResponseDeliverTx) {
	log.Printf("In deliverTx: %s\n", string(req.Tx))

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
		var badge GiveBadge
		if err := json.Unmarshal(payload.Params, &badge); err != nil {
			res.Code = code.CodeTypeUnmarshalError
			res.Log = "error when unmarshal params"
			break
		}

		badge.Giver = nil
		badgeKey, err := json.Marshal(badge)
		if err != nil {
			res.Code = code.CodeTypeEncodingError
			res.Log = "error when marshal badgeKey"
			break
		}

		app.state.currentBatch.Set(badgeKey, []byte(payload.Params))
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
func (app *SitcomApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	log.Printf("In checkTx: %s\n", string(req.Tx))

	var txObj protoTm.Tx
	if err := proto.Unmarshal(req.Tx, &txObj); err != nil {
		log.Println("Error in unmarshal txObj")
		return types.ResponseCheckTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	log.Printf("txObj: %+v\n", txObj)

	payload := txObj.Payload
	pubKey := txObj.PublicKey
	signature := txObj.Signature

	if payload.Method == "" {
		log.Println("Empty method")
		return types.ResponseCheckTx{
			Code: code.CodeTypeEmptyMethod,
			Log:  "method cannot be empty"}
	}

	if _, exists := methodList[payload.Method]; !exists {
		log.Printf("Unknown method %s", payload.Method)
		return types.ResponseCheckTx{
			Code: code.CodeTypeInvalidMethod,
			Log:  fmt.Sprintf("unknown for method %s", payload.Method)}
	}
	/*
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error in marshal payload: %+v\n", payload)
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  "error with payload unmarshal"}
		}
	*/
	log.Printf("pubkey: 0x%x\nparams: %s\nsignature: 0x%x\n", pubKey, payload.Params, signature)

	if !ed25519.Verify(pubKey, payload.Params, signature) {
		log.Printf("Failed in signature verification\n")
		return types.ResponseCheckTx{
			Code: code.CodeTypeUnauthorized,
			Log:  "failed in signature verification",
		}
	}

	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

// Commit commit a current transaction batch
func (app *SitcomApplication) Commit() (res types.ResponseCommit) {
	appHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(appHash, app.state.Size)

	app.state.AppHash = appHash

	app.state.Height++
	app.state.SaveState()
	app.state.currentBatch.Commit()

	res.Data = appHash

	return
}

// Query return data from blockchain
func (app *SitcomApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	log.Printf("In query: %s\n", string(req.Data))

	// For query
	res.Key = req.Data
	parts := strings.Split(string(res.Key), "=")
	if len(parts) == 2 {
		result := make([]byte, 0)
		err := app.state.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			itr := txn.NewIterator(opts)
			defer itr.Close()
			for itr.Rewind(); itr.Valid(); itr.Next() {
				item := itr.Item()
				k := item.Key()
				if bytes.Contains(k, []byte(parts[0])) && bytes.Contains(k, []byte(parts[1])) {
					result = append(result, []byte("|")...)
					result = append(result, k...)
				}
			}
			return nil
		})

		if err != nil {
			res.Log = err.Error()
			res.Code = code.CodeTypeUnknownError
			return
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
	err := app.state.db.View(func(txn *badger.Txn) error {
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
		res.Log = err.Error()
		res.Code = code.CodeTypeUnknownError
		return
	}

	return
}

// InitChain is used for initialize a blockchain
func (app *SitcomApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

// BeginBlock create new transaction batch
func (app *SitcomApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.state.currentBatch = app.state.db.NewTransaction(true)
	return types.ResponseBeginBlock{}
}

// EndBlock is called when ending block
func (app *SitcomApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}
