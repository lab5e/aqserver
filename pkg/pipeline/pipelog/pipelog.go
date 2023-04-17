// Package pipelog implements the log step of the pipeline
package pipelog

import (
	"log"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/pipeline"
)

// Log is a pipeline processor that logs incoming messages
type Log struct {
	next pipeline.Pipeline
}

// New creates new instance of Log pipeline element
func New() *Log {
	return &Log{}
}

// Publish ...
func (p *Log) Publish(m *model.Message) error {
	log.Printf("Message: device='%s' messageID=%d packetSize=%d", m.DeviceID, m.ID, m.PacketSize)

	if p.next != nil {
		return p.next.Publish(m)
	}
	return nil
}

// AddNext ...
func (p *Log) AddNext(pe pipeline.Pipeline) {
	p.next = pe
}

// Next ...
func (p *Log) Next() pipeline.Pipeline {
	return p.next
}
