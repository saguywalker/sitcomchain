package app

import (
	"encoding/binary"
	"fmt"

	"github.com/saguywalker/sitcomchain/code"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

// Info set an information of blockchain
func (app *SITComApplication) Info(req types.RequestInfo) types.ResponseInfo {
	res := types.ResponseInfo{
		Data:       fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:    version.ABCIVersion,
		AppVersion: protocolVersion.Uint64(),
	}
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	fmt.Printf("%+v\n", res)
	return res
}

// DeliverTx updates new data into blockchain
// val:${pubkey}!{0 or 1} => update validator
func (app *SITComApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	tx := req.Tx

	if isValidatorTx(tx) {
		return app.execValidatorTx(tx)
	}

	app.state.db.Set(tx, tx)
	app.state.Size++

	events := []types.Event{
		{
			Type: "blockchain.update",
			Attributes: []cmn.KVPair{
				{Key: []byte("merkle_root"), Value: tx},
			},
		},
	}

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Events: events}
}

// CheckTx validates the tx payload
func (app *SITComApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	if len(req.Tx) != 32 || isValidatorTx(req.Tx) {
		return types.ResponseCheckTx{Code: code.CodeTypeBadData, Log: "tx's size should be 32 or val:${pubkey}!{0 or 1}"}
	}

	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

// Commit updates new current state into blockchain
func (app *SITComApplication) Commit() types.ResponseCommit {
	appHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

// Query value from a corresponding key
func (app *SITComApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	if len(req.Data) == 0 {
		app.state.db.Print()
	} else {
		res.Key = req.Data
		value := app.state.db.Get(req.Data)
		res.Value = value
		if value != nil {
			res.Log = "exists"
		} else {
			res.Log = "does not exist"
		}
	}
	return
}

// InitChain initializes blockchain with specified validator set
func (app *SITComApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// BeginBlock decreases voting power from ByzantineValidators
func (app *SITComApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]types.ValidatorUpdate, 0)

	for _, ev := range req.ByzantineValidators {
		if ev.Type == tmtypes.ABCIEvidenceTypeDuplicateVote {
			// decrease voting power by 1
			if ev.TotalVotingPower == 0 {
				continue
			}
			app.updateValidator(types.ValidatorUpdate{
				PubKey: app.valAddrToPubKeyMap[string(ev.Validator.Address)],
				Power:  ev.TotalVotingPower - 1,
			})
		}
	}
	return types.ResponseBeginBlock{}
}

// EndBlock computes when ending the current block
func (app *SITComApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}
