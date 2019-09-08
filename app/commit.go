package app

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/saguywalker/sitcomchain/code"
	"github.com/saguywalker/sitcomchain/model"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// Validators updates validator
func (app *SITComApplication) Validators() (validators []types.ValidatorUpdate) {
	itr := app.state.db.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
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

func (app *SITComApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
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

func (app *SITComApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
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

// StaffAddCompetence method stores data into blockchain
func (app *SITComApplication) StaffAddCompetence(body []byte) ([]types.Event, error) {
	var update model.StaffAddCompetence
	err := json.Unmarshal(body, &update)
	if err != nil {
		return nil, err
	}

	update.Nonce = app.state.Size + 1

	value, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	semester := make([]byte, 2)
	binary.LittleEndian.PutUint16(semester, update.Semester)

	competenceID := make([]byte, 2)
	binary.LittleEndian.PutUint16(competenceID, update.CompetenceID)

	key := crypto.Sha256(value)

	//Set struct_id to value
	app.state.db.Set(key, value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "competence.add",
			Attributes: []cmn.KVPair{
				{Key: []byte("txid"), Value: key},
				{Key: []byte("studentid"), Value: []byte(update.StudentID)},
				{Key: []byte("competenceid"), Value: competenceID},
				{Key: []byte("by"), Value: []byte(update.By)},
				{Key: []byte("semester"), Value: semester},
			},
		},
	}

	return events, nil
}

// AttendedActivity method stores data into blockchain
func (app *SITComApplication) AttendedActivity(body []byte) ([]types.Event, error) {
	var update model.AttendedActivity
	err := json.Unmarshal(body, &update)
	if err != nil {
		return nil, err
	}

	update.Nonce = app.state.Size + 1

	value, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	semester := make([]byte, 2)
	binary.LittleEndian.PutUint16(semester, update.Semester)

	activityID := make([]byte, 4)
	binary.LittleEndian.PutUint32(activityID, update.ActivityID)

	key := crypto.Sha256(value)

	app.state.db.Set(key, value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "activity.approve",
			Attributes: []cmn.KVPair{
				{Key: []byte("txid"), Value: key},
				{Key: []byte("studentid"), Value: []byte(update.StudentID)},
				{Key: []byte("activityid"), Value: activityID},
				{Key: []byte("by"), Value: update.Approver},
				{Key: []byte("semester"), Value: semester},
			},
		},
	}

	return events, nil
}
