package app

import (
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

// SitcomApplication struct
type SitcomApplication struct {
	types.BaseApplication

	state              State
	valUpdates         map[string]types.ValidatorUpdate
	verifiedSignatures map[string]string
}

var (
	_          types.Application = (*SitcomApplication)(nil)
	methodList                   = map[string]bool{
		"SetValidator": true,
		"GiveBadge":    true,
	}
)

// NewSitcomApp return new SitcomApplication struct with db
func NewSitcomApp(dbDir string) *SitcomApplication {
	db, err := dbm.NewGoLevelDB("sitcomchain", dbDir)
	if err != nil {
		panic(err)
	}
	appState := NewAppState(db)

	return &SitcomApplication{
		state:              appState,
		valUpdates:         make(map[string]types.ValidatorUpdate),
		verifiedSignatures: make(map[string]string),
	}
}
