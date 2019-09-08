package app

import (
	"encoding/json"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"
	dbm "github.com/tendermint/tm-db"
)

var (
	_               types.Application = (*SITComApplication)(nil)
	protocolVersion version.Protocol  = 0x1
	stateKey                          = []byte("stateKey:")
)

// SITComApplication defines an application struct
type SITComApplication struct {
	types.BaseApplication

	state              State
	ValUpdates         []types.ValidatorUpdate
	valAddrToPubKeyMap map[string]types.PubKey
	logger             log.Logger
}

// State defines a struct which contain the current status right now
type State struct {
	db      dbm.DB
	Size    uint64 `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

// NewSITComApplication returns new SITComApplication struct
func NewSITComApplication(dbDir string) *SITComApplication {
	name := "sitcomchain"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &SITComApplication{
		state:              state,
		valAddrToPubKeyMap: make(map[string]types.PubKey),
		logger:             log.NewNopLogger()}
}

// SetLogger sets a logger from log.Logger
func (app *SITComApplication) SetLogger(l log.Logger) {
	app.logger = l
}
