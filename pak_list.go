package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

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
	mp := index.MountPoint()

	for _, r := range index.IndexRecords() {
		fmt.Printf("%s\n", path.Join(mp, r.FileName()))
	}
}
