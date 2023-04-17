// Package pipeline defines the data pipeline interface.
package pipeline

import (
	"github.com/lab5e/aqserver/pkg/model"
)

// Pipeline defines the interface of processing pipeline elements.
type Pipeline interface {
	Publish(m *model.Message) error
	AddNext(pe Pipeline)
	Next() Pipeline
}
