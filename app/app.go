package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/saguywalker/sitcomchain/code"
)

type SitcomApplication struct {
	types.BaseApplication

	db           *badger.DB
	currentBatch *badger.Txn
}

var _ types.Application = (*SitcomApplication)(nil)

func NewSitcomApp(db *badger.DB) *SitcomApplication {
	return &SitcomApplication{
		db: db,
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

func (a *SitcomApplication) Validators() (validators []types.ValidatorUpdate) {
	err := a.db.View(func(txn *badger.Txn) error {
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()

		for ; itr.Valid(); itr.Next() {
			item := itr.Item()
			key := item.Key()
			if isValidatorTx(key) {
				validator := new(types.ValidatorUpdate)
				err := item.Value(func(v []byte) error {
					if err := types.ReadMessage(bytes.NewBuffer(v), validator); err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
				validators = append(validators, *validator)
			}
		}
	})

	if err != nil {
		panic(err)
	}

	return
}

// MakeValSetChangeTx encode base64 and return byte array
func MakeValSetChangeTx(pubkey types.PubKey, power int64) []byte {
	pubStr := base64.StdEncoding.EncodeToString(pubkey.Data)
	return []byte(fmt.Sprintf("val:%s!%d", pubStr, power))
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), "val:")
}

func (app *SitcomApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len("val:"):]

	//get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "!")
	if len(pubKeyAndPower) != 2 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey!power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid base64", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(types.Ed25519ValidatorUpdate(pubkey, int64(power)))
}

func (app *SitcomApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	key := []byte("val:" + string(v.PubKey.Data))

	pubkey := ed25519.PubKeyEd25519{}
	copy(pubkey[:], v.PubKey.Data)

	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			pubStr := base64.StdEncoding.EncodeToString(v.PubKey.Data)
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %s", pubStr)}
		}
		app.state.db.Delete(key)
		delete(app.valAddrToPubKeyMap, string(pubkey.Address()))
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
		app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return types.ResponseDeliverTx{Code: code.CodeTypeOK}
}
