package main

import (
	"gitlab.com/sit-competence/sitcomchain/sitcomapp"
	"os"

	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
)

func main() {
	initSitcomApp()
}

func initSitcomApp() error {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var app types.Application

	app = sitcomapp.NewSITComApplication("sitcom")
	app.(*sitcomapp.SITComApplication).SetLogger(logger.With("module", "sitcomchain"))

	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
	if err != nil {
		return err
	}

	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		return err
	}

	cmn.TrapSignal(logger, func() {
		srv.Stop()
	})

	select {}

}
