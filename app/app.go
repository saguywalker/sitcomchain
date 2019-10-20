package app

import (
	"bytes"

	"github.com/dgraph-io/badger"
	"github.com/tendermint/tendermint/abci/types"

	"github.com/saguywalker/sitcomchain/code"
)

type SitcomApplication struct {
	types.BaseApplication

	state              State
	valUpdates         map[string]types.ValidatorUpdate
	verifiedSignatures map[string]string
}

var _ types.Application = (*SitcomApplication)(nil)

func NewSitcomApp(db *badger.DB) *SitcomApplication {
	appState := NewState(db)

	return &SitcomApplication{
		state:              appState,
		valUpdates:         make(map[string]types.ValidatorUpdate),
		verifiedSignatures: make(map[string]string),
	}
}

func (a *SitcomApplication) isValid(tx []byte) (errCode uint32) {
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return code.CodeTypeEncodingError
	}

	key, value := parts[0], parts[1]

	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					errCode = code.CodeTypeDuplicateKey
				}
				return nil
			})
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return errCode
}
