package spanlistener

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/lab5e/aqserver/pkg/listener"
	"github.com/lab5e/aqserver/pkg/pipeline"
	"github.com/lab5e/aqserver/pkg/util"
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

// New creates a new Listener instance which connects to Span.
func New(pipeline pipeline.Pipeline, apiToken string, collectionID string) listener.Listener {
	return &spanListener{
		pipeline:         pipeline,
		apiToken:         apiToken,
		collectionID:     collectionID,
		apiEndpointURL:   util.EndpointURL(util.DefaultSpanWebsocketEndpointBaseURL, collectionID),
		handshakeTimeout: util.DefaultSpanWebsocketHandshakeTimeout,
		reconnectDelay:   util.DefaultSpanWebsocketReconnectDelay,
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
	s.apiEndpointURL = util.EndpointURL(baseURL, s.collectionID)
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
			decodedMessage, err := util.DecodePayload(message)
			if err != nil {
				if err == util.ErrMessageWasKeepalive {
					continue
				}

				if err == util.ErrMessageWasNotData {
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
