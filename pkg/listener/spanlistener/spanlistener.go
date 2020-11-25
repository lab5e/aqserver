package spanlistener

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lab5e/aqserver/pkg/aqpb"
	"github.com/lab5e/aqserver/pkg/listener"
	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/spanclient-go"
	"google.golang.org/protobuf/proto"
)

type spanListener struct {
	pipeline         pipeline.Pipeline
	apiToken         string
	collectionID     string
	apiEndpointURL   string
	streamCancelFunc context.CancelFunc
	handshakeTimeout time.Duration
	reconnectDelay   time.Duration
	shutdownChan     chan interface{}
}

const (
	defaultAPIEndpointBaseURL = "wss://api.lab5e.com/span"
	defaultHandshakeTimeout   = 50 * time.Second
	defaultReconnectDelay     = 5 * time.Second
)

var (
	// ErrMessageWasKeepalive indicates that the message received was a keepalive
	ErrMessageWasKeepalive = errors.New("Message was keepalive")

	// ErrMessageWasNotData indicates that the messge received was something other than data or keepalive
	ErrMessageWasNotData = errors.New("Message was not data")
)

// New creates a new Listener instance which connects to Span.
func New(pipeline pipeline.Pipeline, apiToken string, collectionID string) listener.Listener {
	return &spanListener{
		pipeline:         pipeline,
		apiToken:         apiToken,
		collectionID:     collectionID,
		apiEndpointURL:   endpointURL(defaultAPIEndpointBaseURL, collectionID),
		handshakeTimeout: defaultHandshakeTimeout,
		reconnectDelay:   defaultReconnectDelay,
		shutdownChan:     make(chan interface{}),
	}
}

// Start starts the listener
func (s *spanListener) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	s.streamCancelFunc = cancelFunc

	go s.dataStreamLoop(ctx)
	return nil
}

func (s *spanListener) Shutdown() {
	s.streamCancelFunc()
}

func (s *spanListener) WaitForShutdown() {
	<-s.shutdownChan
}

func (s *spanListener) SetEndpointBaseURLDebug(baseURL string) {
	s.apiEndpointURL = endpointURL(baseURL, s.collectionID)
}

func (s *spanListener) dataStreamLoop(ctx context.Context) {
	// When we exit the loop we close this channel to signal that
	// there will be no more messages.
	defer close(s.shutdownChan)

	dialer := websocket.Dialer{
		HandshakeTimeout: s.handshakeTimeout,
	}

	header := http.Header{
		"X-API-Token": []string{s.apiToken},
	}

	for {
		conn, response, err := dialer.Dial(s.apiEndpointURL, header)
		if err != nil {
			log.Printf("failed to connect to '%s'. response='%v', error='%s'", s.apiEndpointURL, response, err)
			log.Printf("waiting %s before attempting reconnect", s.reconnectDelay)

			select {
			case <-ctx.Done():
				log.Printf("SpanListener closed: %v", ctx.Err())
				return

			case <-time.After(s.reconnectDelay):
				continue
			}
		}
		log.Printf("connected to %s", s.apiEndpointURL)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("error reading message from Span: %v", err)
				break
			}

			// Ignore anything that isn't message type 1 for now
			if messageType != 1 {
				log.Printf("message type is %d, ignoring", messageType)
				continue
			}

			// Ignore keepalives
			decodedMessage, err := decodePayload(message)
			if err != nil {
				if err == ErrMessageWasKeepalive {
					continue
				}

				if err == ErrMessageWasNotData {
					log.Printf("unknown message, skipping")
					continue
				}

				log.Printf("error decoding payload: %v", err)
				continue
			}

			// Publish the message to the processing pipeline
			s.pipeline.Publish(decodedMessage)

			select {
			case <-ctx.Done():
				log.Printf("spangw closed: %v", ctx.Err())
				return
			default:
				continue
			}
		}
	}
}

// decodePayload peels off layers of protocol to reveal the golden
// nugget that is the sensor data message.
func decodePayload(rawPayload []byte) (*model.Message, error) {
	// Parse JSON
	var outputDataMessage = spanclient.OutputDataMessage{}
	err := json.Unmarshal(rawPayload, &outputDataMessage)
	if err != nil {
		return &model.Message{}, fmt.Errorf("JSON decode failed %w", err)
	}

	if outputDataMessage.Type == "keepalive" {
		return &model.Message{}, ErrMessageWasKeepalive
	}

	if outputDataMessage.Type != "data" {
		return &model.Message{}, ErrMessageWasNotData
	}

	// Decode base64
	bytePayload, err := base64.StdEncoding.DecodeString(outputDataMessage.Payload)
	if err != nil {
		return &model.Message{}, fmt.Errorf("base64 decode failed: %w", err)
	}

	// Decode Protobuffer
	var sample = &aqpb.Sample{}
	err = proto.Unmarshal(bytePayload, sample)
	if err != nil {
		return &model.Message{}, fmt.Errorf("Protobuf decode failed: %w", err)
	}

	// Convert to model.Message and put in packet size and timestamp
	msg := model.MessageFromProtobuf(sample)
	msg.PacketSize = len(bytePayload)
	msg.ReceivedTime, err = strconv.ParseInt(outputDataMessage.Received, 10, 64)
	if err != nil {
		msg.ReceivedTime = time.Now().UnixNano() / int64(time.Millisecond)
	}

	return msg, nil
}

// endpointURL generates a Websocket endpoint address given the base
// URL and a collection id.
func endpointURL(baseURL string, collectionID string) string {
	return fmt.Sprintf("%s/collections/%s/from", baseURL, collectionID)
}
