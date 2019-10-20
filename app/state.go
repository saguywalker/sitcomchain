package app

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
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

	db           *badger.DB
	currentBatch *badger.Txn
	Size         uint64
}

// NewAppState create new state struct
func NewAppState(db *badger.DB) State {
	stateMetaData := loadAppState(db)

	state := State{
		StateMetaData: stateMetaData,
		db:            db,
	}

	return state
}

func loadAppState(db *badger.DB) (stateMetaData StateMetaData) {
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(stateKey)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &stateMetaData); err != nil {
				return err
			}

			return nil
		})

		return err
	})

	if err != nil && err != badger.ErrKeyNotFound {
		panic(err)
	}

	return stateMetaData
}

// SaveState save current state to blockchain
func (state *State) SaveState() {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	state.currentBatch.Set(stateKey, stateBytes)
}
