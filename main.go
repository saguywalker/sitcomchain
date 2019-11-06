package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	abciserver "github.com/tendermint/tendermint/abci/server"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	sitcomapp "github.com/saguywalker/sitcomchain/app"
)

var socketAddr string

func init() {
	flag.StringVar(&socketAddr, "socket-addr", "tcp://0.0.0.0:26658", "socket address")
}

func main() {
	logger := logrus.New()
	app := sitcomapp.NewSitcomApp("data", logrus.NewEntry(logger))

	flag.Parse()

	loggerTm := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// server := abciserver.NewSocketServer(socketAddr, app)
	server, err := abciserver.NewServer(socketAddr, "socket", app)
	if err != nil {
		panic(err)
	}

	server.SetLogger(loggerTm)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	defer server.Stop()
	/*
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		os.Exit(0)
	*/

	cmn.TrapSignal(loggerTm, func() {
		server.Stop()
	})

	select {}
}
