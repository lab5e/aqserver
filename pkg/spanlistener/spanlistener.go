package spanlistener

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"

	aqv1 "github.com/lab5e/aqserver/pkg/aq/v1"
	"github.com/lab5e/aqserver/pkg/model"
	"google.golang.org/protobuf/proto"

	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/go-spanapi/v4/apitools"
)

type SpanListener interface {
	WaitForShutdown()
}

type spanListener struct {
	pipeline     pipeline.Pipeline
	ds           apitools.DataStream
	collectionID string
	token        string
	shutdownCh   chan struct{}
}

var (
	ErrPipelineNil = errors.New("pipeline is nil")
)

func Create(pipeline pipeline.Pipeline, apiToken string, collectionID string) (SpanListener, error) {
	if pipeline == nil {
		return nil, ErrPipelineNil
	}

	listener := &spanListener{
		pipeline:     pipeline,
		collectionID: collectionID,
		token:        apiToken,
		shutdownCh:   make(chan struct{}),
	}

	clientID := fmt.Sprintf("aqserver-%d", time.Now().UnixMicro())

	var err error
	listener.ds, err = apitools.NewMQTTStream(
		apitools.WithAPIToken(apiToken),
		apitools.WithCollectionID(collectionID),
		apitools.WithClientID(clientID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to open CollectionDataStream: %v", err)
	}

	go listener.readDataStream()

	return listener, nil
}

func (s *spanListener) WaitForShutdown() {
	<-s.shutdownCh
}

func (s *spanListener) readDataStream() {
	defer func() {
		log.Printf("connection to Span closed")
		s.ds.Close()
		close(s.shutdownCh)
	}()

	var sample aqv1.Sample
	for {
		odm, err := s.ds.Recv()
		if err != nil {
			log.Printf("error reading message: %v", err)
			return
		}

		// We only care about messages containing data
		if *odm.Type != "data" {
			continue
		}

		payload, err := base64.StdEncoding.DecodeString(odm.GetPayload())
		if err != nil {
			log.Printf("payload error: %v", err)
			continue
		}

		err = proto.Unmarshal(payload, &sample)
		if err != nil {
			log.Printf("payload error: %v", err)
		}

		message := model.MessageFromProtobuf(&sample)
		message.MessageID = odm.GetMessageId()
		s.pipeline.Publish(message)
	}
}
