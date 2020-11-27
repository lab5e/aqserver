package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lab5e/aqserver/pkg/aqpb"
	"github.com/lab5e/aqserver/pkg/model"

	"github.com/lab5e/spanclient-go/v4"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultSpanWebsocketEndpointBaseURL is the base URL for Span websocket interface
	DefaultSpanWebsocketEndpointBaseURL = "wss://api.lab5e.com/span"

	// DefaultSpanWebsocketHandshakeTimeout is how long we will wait to complete
	// Websocket handshake with Span
	DefaultSpanWebsocketHandshakeTimeout = 50 * time.Second

	// DefaultSpanWebsocketReconnectDelay is how long we will wait
	// before we attempt to reconnect if we get disconnected from
	// Span.
	//
	// TODO(borud): replace with a progressive (exponential) reconnect
	//              delay
	DefaultSpanWebsocketReconnectDelay = 5 * time.Second
)

var (
	// ErrMessageWasKeepalive indicates that the message received was a keepalive
	ErrMessageWasKeepalive = errors.New("Message was keepalive")

	// ErrMessageWasNotData indicates that the messge received was something other than data or keepalive
	ErrMessageWasNotData = errors.New("Message was not data")
)

// DecodePayload peels off layers of protocol to reveal the golden
// nugget that is the sensor data message.
func DecodePayload(rawPayload []byte) (*model.Message, error) {
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

// EndpointURL generates a Websocket endpoint address given the base
// URL and a collection id.
func EndpointURL(baseURL string, collectionID string) string {
	return fmt.Sprintf("%s/collections/%s/from", baseURL, collectionID)
}
