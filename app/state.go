package app

import (
	"encoding/json"

	dbm "github.com/tendermint/tm-db"
)

var (
	stateKey = []byte("stateKey")
)

// StateMetaData struct
type StateMetaData struct {
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

// State contains current state data
type State struct {
	StateMetaData

	db   dbm.DB
	Size uint64
}

// NewAppState create new state struct
func NewAppState(db dbm.DB) State {
	stateMetaData := loadAppState(db)

	state := State{
		StateMetaData: stateMetaData,
		db:            db,
	}

	return state
}

func loadAppState(db dbm.DB) (stateMetaData StateMetaData) {
	stateBytes := db.Get(stateKey)
	if len(stateBytes) != 0 {
		if err := json.Unmarshal(stateBytes, &stateMetaData); err != nil {
			panic(err)
		}
	}

	return stateMetaData
}

// SaveState save current state to blockchain
func (state *State) SaveState() {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	state.db.Set(stateKey, stateBytes)
}
