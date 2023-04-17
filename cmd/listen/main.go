package main

import (
	"log"
	"os"

	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/spanlistener"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
)

func main() {
	db, err := sqlitestore.New(":memory:")
	if err != nil {
		log.Fatal(err)
	}

	pipeline := pipeline.New(db)
	listener, err := spanlistener.Create(pipeline, os.Getenv("SPAN_API_TOKEN"), "17dh0cf43jg007")
	if err != nil {
		log.Fatal(err)
	}
	listener.WaitForShutdown()
}
