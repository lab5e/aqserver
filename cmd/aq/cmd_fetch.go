package main

import (
	"encoding/base64"
	"log"
	"strconv"
	"time"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/pipeline/calculate"
	"github.com/lab5e/aqserver/pkg/pipeline/persist"
	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
)

// fetchCmd fetches backlog of data
type fetchCmd struct{}

// Execute ...
func (a *fetchCmd) Execute(args []string) error {
	db, err := getDB()
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

	config := spanapi.NewConfiguration()
	client := spanapi.NewAPIClient(config)

	ctx := apitools.ContextWithAuth(opt.SpanAPIToken)

	lastMessageID := ""
	count := 0
	totalCount := 0
	for {
		items, _, err := client.CollectionsApi.ListCollectionData(ctx, opt.SpanCollectionID).
			Offset(lastMessageID).
			Limit(200).
			Execute()
		if err != nil {
			log.Printf("list collections error: %v", err)
			return err
		}

		if len(items.Data) == 0 {
			log.Printf("done, fetched %d data points", totalCount)
			return nil
		}

		for _, item := range items.Data {
			lastMessageID = *item.MessageId
			received, err := strconv.ParseInt(*item.Received, 10, 64)
			if err != nil {
				log.Printf("Error converting timestamp string (%s): %v", *item.Received, err)
				received = timeToMilliseconds(time.Now())
			}

			bytes, err := base64.StdEncoding.DecodeString(*item.Payload)
			if err != nil {
				log.Printf("error base64-decoding payload='%s': %v", *item.Payload, err)
				continue
			}

			pb, err := model.ProtobufFromData(bytes)
			if err != nil {
				log.Printf("error protobuf-decoding payload='%s': %v", *item.Payload, err)
				continue
			}

			message := model.MessageFromProtobuf(pb)
			message.DeviceID = *item.Device.DeviceId
			message.ReceivedTime = received
			message.PacketSize = len(bytes)

			pipelineRoot.Publish(message)
			count++
			totalCount++
		}
		if count >= 1000 {
			count = 0
			lastItem := items.Data[len(items.Data)-1]

			lastTimestamp := ""
			lastReceivedMS, err := strconv.ParseInt(*lastItem.Received, 10, 64)
			if err == nil {
				lastTimestamp = time.UnixMilli(lastReceivedMS).Format(time.RFC3339)
			}
			log.Printf(" - fetched %d (%s)", totalCount, lastTimestamp)
		}

	}
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
