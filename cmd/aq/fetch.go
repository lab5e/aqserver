package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
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
	PageSize int32 `short:"p" long:"page-size" description:"Number of rows to fetch per page" default:"500"`
}

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

	// Set up pipeline
	pipelineRoot := pipeline.New(db)
	pipelineCalc := calculate.New(db)
	pipelinePersist := persist.New(db)

	pipelineRoot.AddNext(pipelineCalc)
	pipelineCalc.AddNext(pipelinePersist)

	// TODO(borud): here be dragons
	configuration := spanclient.NewConfiguration()
	client := spanclient.NewAPIClient(configuration)

	ctx := context.WithValue(context.Background(), spanclient.ContextAPIKey,
		spanclient.APIKey{
			Key:    opt.SpanAPIToken,
			Prefix: "",
		})

	log.Printf("Listing collection")
	totalCount := 0
	batchCount := 0

	beginningOfTime := optional.NewString("1582900000000")

	var lastItem spanclient.OutputDataMessage
	lastTime := timeToMilliseconds(time.Now())
	for {

		items, _, err := client.CollectionsApi.ListCollectionData(ctx, opt.SpanCollectionID, &spanclient.ListCollectionDataOpts{
			Limit: optional.NewInt32(a.PageSize),
			Start: beginningOfTime,
			End:   optional.NewString(fmt.Sprint(lastTime)),
		})
		if err != nil {
			return fmt.Errorf("Unable to list collection '%s': %w", opt.SpanCollectionID, err)
		}

		log.Printf(">> %d | %s", len(items.Data), msToTime(lastTime))

		// Check if we have reached the end
		if len(items.Data) == 0 {
			log.Printf("last batch")
			break
		}

		for _, item := range items.Data {
			// Skip overlap.
			if item.Received == lastItem.Received && item.Payload == lastItem.Payload {
				log.Printf("Skip overlap")
				continue
			}

			lastItem = item
			totalCount++
			batchCount++
			if batchCount == 1000 {
				log.Printf("Fetched %d records", totalCount)
				batchCount = 0
			}

			// // deal with item here
			// // log.Printf("%s %s", item.Received, item.Payload)
			// bytes, err := base64.StdEncoding.DecodeString(item.Payload)
			// if err != nil {
			// 	log.Printf("Failed to decode base64 encoded payload len=%d: %v", len(item.Payload), err)
			// 	continue
			// }

			// pb, err := model.ProtobufFromData(bytes)
			// if err != nil {
			// 	log.Printf("Failed to unmarshal protobuf len=%d: %v", len(bytes), err)
			// 	continue
			// }

			// m := model.MessageFromProtobuf(pb)
			// if m == nil {
			// 	log.Printf("Unable to create Message from protobuf")
			// 	continue
			// }

			// m.DeviceID = item.Device.DeviceId
			// m.ReceivedTime, err = strconv.ParseInt(item.Received, 10, 64)
			// if err != nil {
			// 	m.ReceivedTime = time.Now().UnixNano() / int64(time.Millisecond)
			// }
			// m.PacketSize = len(bytes)

			// pipelineRoot.Publish(m)
		}

		t, err := strconv.ParseInt(items.Data[len(items.Data)-1].Received, 10, 64)
		if err != nil {
			log.Fatalf("Error reading timestamp: %v", err)
		}

		lastTime = t
	}

	log.Printf("Fetched a total of %d messages", totalCount)
	return nil
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func millisecondsToTime(ms int64) time.Time {
	return time.Unix(ms/1000, 0)
}
