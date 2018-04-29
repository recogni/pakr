package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/recogni/pakr/pak"
)

////////////////////////////////////////////////////////////////////////////////

func fatalOnError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	pakFile := os.Args[1]
	bs, err := ioutil.ReadFile(pakFile)
	fatalOnError(err)

	index, err := pak.ParseIndexRecord(bs)

	// Step 3 :: Walk each record.
	for _, r := range index.IndexRecords() {
		fmt.Printf("File: %s :: %#v\n", r.FileName(), r.Metadata())
		fmt.Printf("%#v\n", r.Metadata().Hash())
	}
}
