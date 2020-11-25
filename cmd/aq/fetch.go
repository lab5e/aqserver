package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/antihax/optional"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/pipeline/calculate"
	"github.com/lab5e/aqserver/pkg/pipeline/persist"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
	"github.com/lab5e/spanclient-go/v4"
)

// FetchCommand fetches backlog of data
type FetchCommand struct {
	PageSize int `short:"p" long:"page-size" description:"Number of rows to fetch per page" default:"250"`
}

// For this application we say that time begins on 2020-03-25
var beginningOfTime = int64(1585094400000)

func init() {
	parser.AddCommand(
		"fetch",
		"Fetch historical data",
		"Fetch historical sensor data from Horde server",
		&FetchCommand{})
}

// Execute ...
func (a *FetchCommand) Execute(args []string) error {
	db, err := sqlitestore.New(opt.DBFilename)
	if err != nil {
		log.Fatalf("Unable to open or create database file '%s': %v", opt.DBFilename, err)
	}
	defer db.Close()

	// Load the calibration data from dir to ensure we have latest
	loadCalibrationData(db, opt.CalibrationDataDir)

	data, err := db.ListMessages(0, 1)
	if err != nil {
		log.Fatalf("Unable to list messages: %v", err)
	}

	if len(data) == 1 {
		// I'm assuming we have to add an entire second of data here
		// in order to make up for the API not having millisecond
		// resolution?
		beginningOfTime = data[0].ReceivedTime + 1
		log.Printf("Will fetch back to %s", msToTime(beginningOfTime))
	}

	// Set up pipeline
	pipelineRoot := pipeline.New(db)
	pipelineCalc := calculate.New(db)
	pipelinePersist := persist.New(db)

	pipelineRoot.AddNext(pipelineCalc)
	pipelineCalc.AddNext(pipelinePersist)

	var since = beginningOfTime
	var until = time.Now().UnixNano() / int64(time.Millisecond)

	// TODO(borud): here be dragons
	configuration := spanclient.NewConfiguration()
	configuration.Debug = true
	client := spanclient.NewAPIClient(configuration)

	ctx := context.WithValue(context.Background(), spanclient.ContextAPIKey,
		spanclient.APIKey{
			Key:    opt.SpanAPIToken,
			Prefix: "",
		})

	options := &spanclient.ListCollectionDataOpts{
		Limit: optional.NewInt32(10),
		Start: optional.NewString(fmt.Sprint(since)),
		End:   optional.NewString(fmt.Sprint(until)),
	}

	log.Printf("Listing collection")
	items, _, err := client.CollectionsApi.ListCollectionData(ctx, opt.SpanCollectionID, options)
	if err != nil {
		return fmt.Errorf("Unable to list collection '%s': %w", opt.SpanCollectionID, err)
	}

	for _, item := range items.Data {
		log.Printf("> %+v", item)
	}

	log.Printf("Fetched a total of %d messages", len(items.Data))
	return nil
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
