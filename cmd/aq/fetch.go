package main

import (
	"log"
	"time"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/pipeline/calculate"
	"github.com/lab5e/aqserver/pkg/pipeline/persist"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
	"github.com/telenordigital/nbiot-go"
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
	client, err := nbiot.New()
	if err != nil {
		return err
	}

	db, err := sqlitestore.New(options.DBFilename)
	if err != nil {
		log.Fatalf("Unable to open or create database file '%s': %v", options.DBFilename, err)
	}
	defer db.Close()

	// Load the calibration data from dir to ensure we have latest
	loadCalibrationData(db, options.CalibrationDataDir)

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
	pipelineRoot := pipeline.New(&options, db)
	pipelineCalc := calculate.New(&options, db)
	pipelinePersist := persist.New(&options, db)

	pipelineRoot.AddNext(pipelineCalc)
	pipelineCalc.AddNext(pipelinePersist)

	var since = beginningOfTime
	var until = time.Now().UnixNano() / int64(time.Millisecond)
	var count = 0
	var countTotal = 0

	// CollectionData coming from horde arrives in descending order
	// from Received.  So we have to work our way backwards.  We do
	// this by starting with until being equal to "now" and then set
	// the next until value from the last entry we got.
	for {
		data, err := client.CollectionData(options.HordeCollection, msToTime(since), msToTime(until), a.PageSize)
		if err != nil {
			log.Fatalf("Error while reading data: %v", err)
		}

		if len(data) == 0 {
			break
		}

		for _, d := range data {
			pb, err := model.ProtobufFromData(d.Payload)
			if err != nil {
				log.Printf("Failed to decode protobuffer len=%d: %v", len(d.Payload), err)
				continue
			}

			m := model.MessageFromProtobuf(pb)
			if m == nil {
				log.Printf("Unable to create Message from protobuf")
				continue
			}

			m.DeviceID = d.Device.ID
			m.ReceivedTime = d.Received
			m.PacketSize = len(d.Payload)

			pipelineRoot.Publish(m)
			count++
			countTotal++
		}

		until = data[len(data)-1].Received - 1
		if count >= 500 {
			log.Printf("Imported %d records...", countTotal)
			count = 0
		}
	}

	log.Printf("Fetched a total of %d messages", countTotal)
	return nil
}
