package main

import (
	"context"
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
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
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

const (
	spanMaxBatchSize = 500
)

// Execute ...
func (a *FetchCommand) Execute(args []string) error {
	db, err := sqlitestore.New(opt.DBFilename)
	if err != nil {
		log.Fatalf("Unable to open or create database file '%s': %v", opt.DBFilename, err)
	}
	defer db.Close()

	// Load the calibration data from dir to ensure we have latest
	loadCalibrationData(db, opt.CalibrationDataDir)

	// Find last seen record
	lastReceived := fmt.Sprintf("%d", time.Now().Unix()*1000)

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

	// TODO(borud): here be dragons
	configuration := spanclient.NewConfiguration()
	client := spanclient.NewAPIClient(configuration)

	ctx := context.WithValue(context.Background(), spanclient.ContextAPIKey,
		spanclient.APIKey{
			Key:    opt.SpanAPIToken,
			Prefix: "",
		})

	total := 0
	seenOldest := false

	// last item of next batch is first item of previous batch
	firstItemOfLastBatch := spanclient.OutputDataMessage{}

	for {
		options := &spanclient.ListCollectionDataOpts{
			// Do not set limit
			Start: optional.NewString("0"),
			End:   optional.NewString(lastReceived),
		}

		start := time.Now()
		items, _, err := client.CollectionsApi.ListCollectionData(ctx, opt.SpanCollectionID, options)
		if err != nil {
			return err
		}
		duration := time.Since(start)

		// The oldest item is at the top
		lastReceived = items.Data[0].Received
		total += len(items.Data)

		msTimestamp, err := strconv.ParseInt(lastReceived, 10, 64)
		if err != nil {
			fmt.Println("Error converting timestamp ", lastReceived, " to a number")
		}

		ts := time.Unix(0, msTimestamp*int64(time.Millisecond))
		log.Printf("fetch duration='%s' %d records (%d total). Last timestamp = %s (%s)\n", duration, len(items.Data), total, ts.String(), lastReceived)

		for _, item := range reverse(items.Data) {
			if item.Received == firstItemOfLastBatch.Received && item.Payload == firstItemOfLastBatch.Payload {
				// This skips the overlap.  Yes, it is ugly.
				continue
			}

			received, err := strconv.ParseInt(item.Received, 10, 64)
			if err != nil {
				log.Printf("item without timestamp")
				received = timeToMilliseconds(time.Now())
			}

			if received <= a.StopAt {
				seenOldest = true
				log.Printf("Hit limit at %d", a.StopAt)
				break
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

		// Check if last batch was partially filled or if we have seen
		// the oldest message we're supposed to fetch.
		if seenOldest || len(items.Data) < spanMaxBatchSize {
			break
		}
		firstItemOfLastBatch = items.Data[0]
	}
	fmt.Println("done")

	return nil
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func millisecondsToTime(ms int64) time.Time {
	return time.Unix(ms/1000, 0)
}

func reverse(s []spanclient.OutputDataMessage) []spanclient.OutputDataMessage {
	a := make([]spanclient.OutputDataMessage, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}
