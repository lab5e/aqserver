package hordelistener

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/opts"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/telenordigital/nbiot-go"
)

// ErrNoHordeCollection indicates that the Horde collectionID is an empty string
var ErrNoHordeCollection = errors.New("No Horde CollectionID specified, please set -c or --collection command line option")

// HordeListener connects to Horde and listens for messages from a
// particular Collection
type HordeListener struct {
	pipeline     pipeline.Pipeline
	collectionID string
	doneChan     chan error
	quit         chan bool
	client       *nbiot.Client
	opts         *opts.Opts
}

// New creates a new HordeListener instance
func New(opts *opts.Opts, pipeline pipeline.Pipeline) *HordeListener {
	return &HordeListener{
		opts:         opts,
		pipeline:     pipeline,
		collectionID: opts.HordeCollection,
		doneChan:     make(chan error),
		quit:         make(chan bool),
	}
}

const reconnectDelay = 5 * time.Second

// Start HordeListener instance
func (h *HordeListener) Start() error {
	if h.collectionID == "" {
		return ErrNoHordeCollection
	}

	c, err := nbiot.New()
	if err != nil {
		return err
	}
	h.client = c

	go func() {
		for {
			stream, err := h.client.CollectionOutputStream(h.collectionID)
			if err != nil {
				log.Fatal("Error connecting to Horde: ", err)
			}

			log.Printf("Connected. Starting Horde listening loop")
			for {
				data, err := stream.Recv()
				if err == io.EOF {
					h.doneChan <- err
					break
				}
				if err != nil {
					h.doneChan <- err
				}

				pb, err := model.ProtobufFromData(data.Payload)
				if err != nil {
					log.Printf("Failed to decode protobuffer len=%d: %v", len(data.Payload), err)
					continue
				}

				m := model.MessageFromProtobuf(pb)
				if m == nil {
					log.Printf("Unable to create Message from protobuf")
					continue
				}

				m.DeviceID = data.Device.ID
				m.ReceivedTime = data.Received
				m.PacketSize = len(data.Payload)

				// TODO(borud): This is a good place to check if a device
				//    is already known and inject it into the database.

				if h.opts.Verbose {
					log.Printf("Accepted packet from Horde %v", m)
				}
				h.pipeline.Publish(m)
			}
			log.Printf("Lost connection to Horde.  Will wait for %v to reconnect", reconnectDelay)
			time.Sleep(reconnectDelay)
			log.Printf("Reconnecting to Horde")
		}
	}()
	return nil
}

// Shutdown initiates shutdown of the UDPListener
func (h *HordeListener) Shutdown() {
	log.Printf("UDPListener: Shutdown not implemented")
}

// WaitForShutdown waits for the UDP listener to shut down
func (h *HordeListener) WaitForShutdown() {
	<-h.quit
}
