package main

import (
	"os"

	"github.com/saguywalker/sitcomchain/app"
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
	var a types.Application

	a = app.NewSITComApplication("sitcomdata")
	a.(*app.SITComApplication).SetLogger(logger.With("module", "sitcomchain"))

	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", a)
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
