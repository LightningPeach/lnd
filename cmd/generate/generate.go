package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btclog"
	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lightninglabs/neutrino"
)

const (
	blockHeadersFilename     = "block_headers.bin"
	regFilterHeadersFilename = "reg_filter_headers.bin"
	neutrinoDBFilename       = "neutrino.db"
)

var (
	backendLog = btclog.NewBackend(os.Stdout)

	btcnLog = backendLog.Logger("BTCN")
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "presync generation failed: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	dirHelp := fmt.Sprintf("path to directory with presynced files(%v%v%v)",
		blockHeadersFilename, regFilterHeadersFilename, neutrinoDBFilename)
	dir := flag.String("dir", "neutrino", dirHelp)
	flag.Parse()

	checkErr(os.MkdirAll(*dir, 0755))

	db, err := walletdb.Create("bdb", filepath.Join(*dir, neutrinoDBFilename))
	checkErr(err)

	cfg := neutrino.Config{
		DataDir:         *dir,
		Database:        db,
		ChainParams:     chaincfg.TestNet3Params,
		ConnectPeers:    []string{"testnetwallet.lightningpeach.com:18333"},
		FilterCacheSize: neutrino.DefaultFilterCacheSize,
		BlockCacheSize:  neutrino.DefaultBlockCacheSize,
	}

	cs, err := neutrino.NewChainService(cfg)
	checkErr(err)

	cs.Start()

	done := make(chan struct{}, 0)
	<-done
}

func init() {
	neutrino.UseLogger(btcnLog)
}
