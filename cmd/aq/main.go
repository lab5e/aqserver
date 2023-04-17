package main

import (
	"github.com/lab5e/aqserver/pkg/store"
	"github.com/lab5e/aqserver/pkg/store/mysqlstore"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
	"github.com/lab5e/aqserver/pkg/util"
)

var opt struct {
	DBFilename         string `long:"db" description:"Data storage file for build in database" default:"aq.db" value-name:"<file>"`
	SpanCollectionID   string `long:"span-collection-id" description:"Span collection to listen to" env:"SPAN_COLLECTION_ID" default:"17dh0cf43jg007" value-name:"<collectionID>"`
	SpanAPIToken       string `long:"span-api-token" description:"Span API token" env:"SPAN_API_TOKEN" value-name:"<Span API token>"`
	CalibrationDataDir string `long:"cal-data-dir" description:"Directory where calibration data is picked up" default:"calibration-data" value-name:"<DIR>"`
	// "<username>:<secret>@<host>:<port>)/<database>?parseTime=true"
	MySQLConnectString string `long:"mysql-connect-string" env:"MYSQL_CONNECT_STRING" description:"MySQL connect string"`
	Verbose            bool   `short:"v"`

	Fetch  fetchCmd  `command:"fetch" description:"fetch data backlog"`
	Import importCmd `command:"import" description:"import calibration data"`
	List   listCmd   `command:"list" description:"list calibration data"`
	Server serverCmd `command:"server" description:"run server"`
}

func main() {
	util.FlagParse(&opt)
}

func getDB() (store.Store, error) {
	if opt.MySQLConnectString != "" {
		return mysqlstore.New(opt.MySQLConnectString)
	}
	return sqlitestore.New(opt.DBFilename)
}
