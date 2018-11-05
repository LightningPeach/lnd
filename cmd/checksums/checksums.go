package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	blockHeadersFilename     = "block_headers.bin"
	regFilterHeadersFilename = "reg_filter_headers.bin"
	neutrinoDBFilename       = "neutrino.db"
)

const (
	blockHeadersCheckSum     = "f37fda67ab7a9d4697c8a0b36151fb84222418415bf41c0153ee7cf81716fd50"
	regFilterHeadersCheckSum = "c3220ffc5096c040faa3e23661fcb2d0fe6509dd74a300f99b9d5973b88348dc"
	neutrinoDBCheckSum       = "d3c1f7fd57a18120c24baf672aad549a9ea1ef9e623eb16f5d8886f2aaf030d4"
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "presync verification failed: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	dirHelp := fmt.Sprintf("path to directory with presynced files(%v%v%v)",
		blockHeadersFilename, regFilterHeadersFilename, neutrinoDBFilename)
	dir := flag.String("dir", "neutrino", dirHelp)
	flag.Parse()

	blockHeadersRaw, err := ioutil.ReadFile(filepath.Join(*dir, blockHeadersFilename))
	checkErr(err)
	checkSum := sha256.Sum256(blockHeadersRaw)
	assertEq(hex.EncodeToString(checkSum[:]), blockHeadersCheckSum, "bad block headers's checksum")

	regFilterHeadersRaw, err := ioutil.ReadFile(filepath.Join(*dir, regFilterHeadersFilename))
	checkErr(err)
	checkSum = sha256.Sum256(regFilterHeadersRaw)
	assertEq(hex.EncodeToString(checkSum[:]), regFilterHeadersCheckSum, "bad reg filter headers's checksum")

	neutrinoDBRaw, err := ioutil.ReadFile(filepath.Join(*dir, neutrinoDBFilename))
	checkErr(err)
	checkSum = sha256.Sum256(neutrinoDBRaw)
	assertEq(hex.EncodeToString(checkSum[:]), neutrinoDBCheckSum, "bad neutrino db's checksum")
}

func assertEq(left, right, meta string) {
	if left != right {
		checkErr(errors.New(meta))
	}
}
