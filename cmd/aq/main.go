package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type options struct {
	DBFilename         string `long:"db" description:"Data storage file" default:"aq.db" value-name:"<file>"`
	SpanCollectionID   string `long:"span-collection-id" description:"Span collection to listen to" env:"SPAN_COLLECTION_ID" default:"17dh0cf43jg007" value-name:"<collectionID>"`
	SpanAPIToken       string `long:"span-api-token" description:"Span API token" env:"SPAN_API_TOKEN" value-name:"<Span API token>"`
	CalibrationDataDir string `long:"cal-data-dir" description:"Directory where calibration data is picked up" default:"calibration-data" value-name:"<DIR>"`
	Verbose            bool   `short:"v"`
}

var opt options
var parser = flags.NewParser(&opt, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
