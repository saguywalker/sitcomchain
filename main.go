package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/badger"
	abciserver "github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/libs/log"

	sitcomapp "github.com/saguywalker/sitcomchain/app"
)

var socketAddr string

func init() {
	flag.StringVar(&socketAddr, "socket-addr", "tcp://0.0.0.0:26658", "socket address")
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	db, err := badger.Open(badger.DefaultOptions(fmt.Sprintf("%s/.tendermint/data", homeDir)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	app := sitcomapp.NewSitcomApp(db)

	flag.Parse()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// server := abciserver.NewSocketServer(socketAddr, app)
	server, err := abciserver.NewServer(socketAddr, "socket", app)
	if err != nil {
		panic(err)
	}

	server.SetLogger(logger)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting socket server: %v", err)
		os.Exit(1)
	}
	defer server.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}
