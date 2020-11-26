package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lab5e/aqserver/pkg/store"
	"github.com/lab5e/aqserver/pkg/store/mysqlstore"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
)

type options struct {
	DBFilename         string `long:"db" description:"Data storage file for build in database" default:"aq.db" value-name:"<file>"`
	SpanCollectionID   string `long:"span-collection-id" description:"Span collection to listen to" env:"SPAN_COLLECTION_ID" default:"17dh0cf43jg007" value-name:"<collectionID>"`
	SpanAPIToken       string `long:"span-api-token" description:"Span API token" env:"SPAN_API_TOKEN" value-name:"<Span API token>"`
	CalibrationDataDir string `long:"cal-data-dir" description:"Directory where calibration data is picked up" default:"calibration-data" value-name:"<DIR>"`

	// MySQL connect string.  If this is non-empty we use MySQL rather than the built-in SQLite database
	MySQLConnectString string `long:"mysql-connect-string" env:"MYSQL_CONNECT_STRING" description:"MySQL connect string" value-name:"<username>:<secret>@<host>:<port>)/<database>?parseTime=true"`

	Verbose bool `short:"v"`
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

func getDB() (store.Store, error) {
	if opt.MySQLConnectString != "" {
		return mysqlstore.New(opt.MySQLConnectString)
	}
	return sqlitestore.New(opt.DBFilename)
}
