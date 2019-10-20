package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/tendermint/tendermint/abci/types"

	"github.com/saguywalker/sitcomchain/code"
)

const (
	// ValidatorSetChangePrefix define the prefix in key
	ValidatorSetChangePrefix string = "val:"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *SitcomApplication) Validators() (validators []types.Validator) {
	err := app.state.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		itr := txn.NewIterator(opts)
		defer itr.Close()

		for itr.Rewind(); itr.ValidForPrefix([]byte(ValidatorSetChangePrefix)); itr.Next() {
			item := itr.Item()
			key := item.Key()
			validator := new(types.Validator)
			if err := types.ReadMessage(bytes.NewBuffer(key), validator); err != nil {
				return err
			}

			validators = append(validators, *validator)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return
}

// add, update, or remove a validator
func (app *SitcomApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	pubKeyBase64 := base64.StdEncoding.EncodeToString(v.PubKey.GetData())
	key := []byte("val:" + pubKeyBase64)

	if v.Power == 0 {
		// remove validator
		err := app.state.db.View(func(txn *badger.Txn) error {
			if _, err := txn.Get(key); err != badger.ErrKeyNotFound {
				return err
			}

			return nil
		})

		if err != badger.ErrKeyNotFound {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %x", key)}
		}

		app.state.currentBatch.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.currentBatch.Set(key, value.Bytes())
	}

	app.valUpdates[pubKeyBase64] = v
	return types.ResponseDeliverTx{
		Code: code.CodeTypeOK,
		Log:  "success"}
}

func (app *SitcomApplication) setValidator(param string) types.ResponseDeliverTx {
	var funcParam SetValidatorParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeUnmarshalError,
			Log:  err.Error()}
	}

	pubKey, err := base64.StdEncoding.DecodeString(string(funcParam.PublicKey))
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeDecodingError,
			Log:  err.Error()}
	}
	var pubKeyObj types.PubKey
	pubKeyObj.Type = "ed25519"
	pubKeyObj.Data = pubKey
	var newValidator types.ValidatorUpdate
	newValidator.PubKey = pubKeyObj
	newValidator.Power = funcParam.Power
	return app.updateValidator(newValidator)
}
