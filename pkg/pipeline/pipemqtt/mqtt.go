// Package pipemqtt is a very simplistic example of how to forward messages to
// an MQTT server.  If you want to use this for production you should
// rewrite it for more robustness.
package pipemqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
)

// MQTTStream ...
type MQTTStream struct {
	client      mqtt.Client
	topicPrefix string
	next        pipeline.Pipeline
}

// New ...
func New(clientID string, password string, address string, topicPrefix string) *MQTTStream {
	opts := createCLientOptions(clientID, password, address)
	client := mqtt.NewClient(opts)
	token := client.Connect()

	for !token.WaitTimeout(3 * time.Second) {
		log.Printf("waiting for MQTT broker")
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}

	return &MQTTStream{
		client:      client,
		topicPrefix: topicPrefix,
	}

}

func createCLientOptions(clientID string, password string, address string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", address))
	opts.SetClientID(clientID)
	opts.SetPassword(password)
	return opts
}

// Publish ...
func (p *MQTTStream) Publish(m *model.Message) error {
	json, err := json.Marshal(m)
	if err == nil {
		topic := fmt.Sprintf("%s/%s", p.topicPrefix, m.DeviceID)
		token := p.client.Publish(topic, 0, false, json)
		if token.Error() != nil {
			log.Printf("Error publishing to MQTT: %v", token.Error())
		}

		if !token.WaitTimeout(10 * time.Millisecond) {
			log.Printf("MQTT publish timed out")
		}
	}

	if p.next != nil {
		return p.next.Publish(m)
	}
	return nil
}

// AddNext ...
func (p *MQTTStream) AddNext(pe pipeline.Pipeline) {
	p.next = pe
}

// Next ...
func (p *MQTTStream) Next() pipeline.Pipeline {
	return p.next
}
