package main

import (
	"log"

	"github.com/lab5e/aqserver/pkg/api"
	"github.com/lab5e/aqserver/pkg/listener"
	"github.com/lab5e/aqserver/pkg/listener/spanlistener"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/pipeline/calculate"
	"github.com/lab5e/aqserver/pkg/pipeline/circular"
	"github.com/lab5e/aqserver/pkg/pipeline/persist"
	"github.com/lab5e/aqserver/pkg/pipeline/pipelog"
	"github.com/lab5e/aqserver/pkg/pipeline/pipemqtt"
	"github.com/lab5e/aqserver/pkg/pipeline/stream"
)

const (
	circularBufferLength = 100
)

type serverCmd struct {
	// Webserver options
	WebListenAddr   string `long:"web-listen-address" description:"Listen address for webserver" default:":8888" value-name:"<[host]:port>"`
	WebAccessLogDir string `long:"web-access-log-dir" description:"Directory for access logs" default:"./logs" value-name:"<dir>"`

	// MQTT
	MQTTAddress     string `long:"mqtt-address" description:"MQTT Address" default:"" value-name:"<[host]:port>"`
	MQTTClientID    string `long:"mqtt-client-id" env:"MQTT_CLIENT_ID" description:"MQTT Client ID" default:""`
	MQTTPassword    string `long:"mqtt-password" env:"MQTT_PASSWORD" description:"MQTT Password" default:""`
	MQTTTopicPrefix string `long:"mqtt-topic-prefix" description:"MQTT topic prefix" default:"aq" value-name:"MQTT topic prefix"`

	// UDP listener
	UDPListenAddress string `long:"udp-listener" description:"Listen address for UDP listener" default:"" value-name:"<[host]:port>"`
	UDPBufferSize    int    `long:"udp-buffer-size" description:"Size of UDP read buffer" default:"1024" value-name:"<num bytes>"`
}

var listeners []listener.Listener

func (a *serverCmd) startSpanListener(r pipeline.Pipeline) listener.Listener {
	log.Printf("Starting Span listener, listening to collection='%s'", opt.SpanCollectionID)
	spanListener := spanlistener.New(r, opt.SpanAPIToken, opt.SpanCollectionID)
	err := spanListener.Start()
	if err != nil {
		log.Fatalf("Unable to start Span listener: %v", err)
	}
	listeners = append(listeners, spanListener)
	return spanListener
}

// Execute ...
func (a *serverCmd) Execute(args []string) error {
	// Set up persistence
	db, err := getDB()
	if err != nil {
		log.Fatalf("Unable to open or create database file '%s': %v", opt.DBFilename, err)
	}
	defer db.Close()

	// Load the calibration data to pick up any new calibration sets.
	err = loadCalibrationData(db, opt.CalibrationDataDir)
	if err != nil {
		// At this point we don't actually care if this returns an
		// error because it just means that we won't get any new
		// calibration data that might have been placed there.
		log.Printf("Did not load any (new) calibration data: %v", err)
	}

	// Create pipeline elements
	// TODO(borud): make streaming broker configurable
	pipelineRoot := pipeline.New(db)
	pipelineCalc := calculate.New(db)
	pipelinePersist := persist.New(db)
	pipelineLog := pipelog.New()
	pipelineCirc := circular.New(circularBufferLength)
	pipelineStream := stream.NewBroker()

	// Chain them together
	pipelineRoot.AddNext(pipelineCalc)
	pipelineCalc.AddNext(pipelinePersist)
	pipelinePersist.AddNext(pipelineLog)
	pipelineLog.AddNext(pipelineStream)
	pipelineStream.AddNext(pipelineCirc)

	// Stream to MQTT server if enabled
	if a.MQTTAddress != "" {
		pipelineMQTT := pipemqtt.New(a.MQTTClientID, a.MQTTPassword, a.MQTTAddress, a.MQTTTopicPrefix)
		pipelineCirc.AddNext(pipelineMQTT)
	}

	// Start Horde listener if enabled
	a.startSpanListener(pipelineRoot)

	// If we have no listeners there is no point to starting so we terminate
	if len(listeners) == 0 {
		log.Fatalf("No listeners defined so terminating.  Please specify at least one listener.")
	}

	// Start api server
	api := api.New(&api.ServerConfig{
		Broker:         pipelineStream,
		DB:             db,
		CircularBuffer: pipelineCirc,
		ListenAddr:     a.WebListenAddr,
		AccessLogDir:   a.WebAccessLogDir,
	})
	api.Start()

	// Wait for all listeners to shut down
	for _, listener := range listeners {
		listener.WaitForShutdown()
	}
	api.Shutdown()
	return nil
}
