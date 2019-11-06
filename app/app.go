package app

import (
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/saguywalker/sitcomchain/version"
)

// SitcomApplication struct
type SitcomApplication struct {
	types.BaseApplication

	AppProtocolVersion uint64
	CurrentChain       string
	Version            string
	logger             *logrus.Entry
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
func NewSitcomApp(dbDir string, logger *logrus.Entry) *SitcomApplication {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorln(r)
		}
	}()

	db, err := dbm.NewGoLevelDB("sitcomchain", dbDir)
	if err != nil {
		panic(err)
	}
	appState := NewAppState(db)

	return &SitcomApplication{
		AppProtocolVersion: version.AppProtocolVersion,
		Version:            version.Version,
		state:              appState,
		valUpdates:         make(map[string]types.ValidatorUpdate),
		verifiedSignatures: make(map[string]string),
	}
}
