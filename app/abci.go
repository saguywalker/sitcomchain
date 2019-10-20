package app

import (
	"bytes"

	"github.com/dgraph-io/badger"
	"github.com/tendermint/tendermint/abci/types"
)

func (a *SitcomApplication) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{}
}

func (a *SitcomApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return types.ResponseSetOption{}
}

func (a *SitcomApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	code := a.isValid(req.Tx)
	if code != 0 {
		return types.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]

	if err := a.currentBatch.Set(key, value); err != nil {
		panic(err)
	}

	return types.ResponseDeliverTx{Code: 0}
}

func (a *SitcomApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	code := a.isValid(req.Tx)
	return types.ResponseCheckTx{Code: code}
}

func (a *SitcomApplication) Commit() types.ResponseCommit {
	a.currentBatch.Commit()
	return types.ResponseCommit{Data: []byte{}}
}

func (a *SitcomApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	res.Key = req.Data
	err := a.db.View(func(txn *badger.Txn) error {
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

func (a *SitcomApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	return types.ResponseInitChain{}
}

func (a *SitcomApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	a.currentBatch = a.db.NewTransaction(true)
	return types.ResponseBeginBlock{}
}

func (a *SitcomApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{}
}
