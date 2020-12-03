package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/pipeline/calculate"
	"github.com/lab5e/aqserver/pkg/pipeline/persist"
	"github.com/lab5e/spanclient-go/v4"
)

// FetchCommand fetches backlog of data
type FetchCommand struct {
	StopAt int64 `long:"stop-at" description:"Timestamp of oldest message we should fetch" default:"0"`
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
	db, err := getDB()
	if err != nil {
		log.Fatalf("Unable to open or create database file '%s': %v", opt.DBFilename, err)
	}
	defer db.Close()

	// Load the calibration data from dir to ensure we have latest
	loadCalibrationData(db, opt.CalibrationDataDir)

	if a.StopAt == 0 {
		lastMessage, err := db.ListMessages(0, 1)
		if err != nil {
			log.Fatalf("Unable to list messages: %v", err)
		}

		if len(lastMessage) == 1 {
			// I'm assuming we have to add an entire second of data here
			// in order to make up for the API not having millisecond
			// resolution?
			a.StopAt = lastMessage[0].ReceivedTime
		}
	}
	log.Printf("Will fetch back to %d", a.StopAt)

	// Set up pipeline
	pipelineRoot := pipeline.New(db)
	pipelineCalc := calculate.New(db)
	pipelinePersist := persist.New(db)

	pipelineRoot.AddNext(pipelineCalc)
	pipelineCalc.AddNext(pipelinePersist)

	configuration := spanclient.NewConfiguration()
	client := spanclient.NewAPIClient(configuration)

	ctx := spanclient.NewAuthContext(opt.SpanAPIToken)

	total := 0

	lastMessageID := ""
	for {
		options := &spanclient.ListCollectionDataOpts{
			// Do not set limit
			Limit:  optional.NewInt32(500),
			Offset: optional.NewString(lastMessageID),
		}

		start := time.Now()
		items, _, err := client.CollectionsApi.ListCollectionData(ctx, opt.SpanCollectionID, options)
		if err != nil {
			return err
		}
		duration := time.Since(start)

		// The oldest item is at the top
		total += len(items.Data)

		if len(items.Data) == 0 {
			// No more data
			fmt.Println("done")
			return nil
		}
		log.Printf("fetch duration='%s' %d records (%d total). Last offset = %s\n", duration, len(items.Data), total, lastMessageID)

		for _, item := range items.Data {
			lastMessageID = item.MessageId
			received, err := strconv.ParseInt(item.Received, 10, 64)
			if err != nil {
				log.Printf("Error converting timestamp string (%s): %v", item.Received, err)
				received = timeToMilliseconds(time.Now())
			}
			if received < a.StopAt {
				// Skip through the remaining items
				fmt.Println("done")
				return nil
			}

			bytes, err := base64.StdEncoding.DecodeString(item.Payload)
			if err != nil {
				log.Printf("error base64-decoding payload='%s': %v", item.Payload, err)
				continue
			}

			pb, err := model.ProtobufFromData(bytes)
			if err != nil {
				log.Printf("error protobuf-decoding payload='%s': %v", item.Payload, err)
				continue
			}

			message := model.MessageFromProtobuf(pb)
			message.DeviceID = item.Device.DeviceId
			message.ReceivedTime = received
			message.PacketSize = len(bytes)

			pipelineRoot.Publish(message)
		}
	}
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
