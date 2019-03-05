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
	defaultBlockHeadersCheckSum     = "9cd44bce499a85763a273c62ba8787fa298edd31d377dd14fa387d79bb6ca8ea"
	defaultRegFilterHeadersCheckSum = "b993ae7440443b7caaf19ecd5269b97180066c3fca393e9556b4bf74642a7cd8"
	defaultNeutrinoDBCheckSum       = "985db27d6ead3849d061bbe82512a2aacd2101d87509fcf1d4705e3ff8e7106c"
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
	show := flag.Bool("show", false, "generate and show checksums")

	blockHeadersCheckSum := flag.String("block_headers_check_sum", defaultBlockHeadersCheckSum, "")
	regFilterHeadersCheckSum := flag.String("reg_filter_headers_check_sum", defaultRegFilterHeadersCheckSum, "")
	neutrinoDBCheckSum := flag.String("neutrino_db_check_sum", defaultNeutrinoDBCheckSum, "")

	flag.Parse()

	blockHeadersRaw, err := ioutil.ReadFile(filepath.Join(*dir, blockHeadersFilename))
	checkErr(err)
	checkSum := sha256.Sum256(blockHeadersRaw)
	checkSumHex := hex.EncodeToString(checkSum[:])
	if *show {
		fmt.Println(checkSumHex)
	} else {
		assertEq(checkSumHex, *blockHeadersCheckSum,
			fmt.Sprintf("bad block headers's checksum; actual: %v, want: %v", checkSumHex, blockHeadersCheckSum))
	}


	regFilterHeadersRaw, err := ioutil.ReadFile(filepath.Join(*dir, regFilterHeadersFilename))
	checkErr(err)
	checkSum = sha256.Sum256(regFilterHeadersRaw)
	checkSumHex = hex.EncodeToString(checkSum[:])
	if *show {
		fmt.Println(checkSumHex)
	} else {
		assertEq(checkSumHex, *regFilterHeadersCheckSum,
			fmt.Sprintf("bad reg filter headers's checksum; actual: %v, want: %v", checkSumHex, regFilterHeadersCheckSum))
	}

	neutrinoDBRaw, err := ioutil.ReadFile(filepath.Join(*dir, neutrinoDBFilename))
	checkErr(err)
	checkSum = sha256.Sum256(neutrinoDBRaw)
	checkSumHex = hex.EncodeToString(checkSum[:])
	if *show {
		fmt.Println(checkSumHex)
	} else {
		assertEq(checkSumHex, *neutrinoDBCheckSum,
			fmt.Sprintf("bad neutrino db's checksum; actual: %v, want: %v", checkSumHex, neutrinoDBCheckSum))
	}
	fmt.Println("OK")
}

func assertEq(left, right, meta string) {
	if left != right {
		checkErr(errors.New(meta))
	}
}
