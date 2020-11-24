package pipeline

import (
	"errors"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/opts"
	"github.com/lab5e/aqserver/pkg/store"
)

// ErrEmptyPipeline indicates that the Root pipeline has no next
// element.  Which makes it kind of useless.
var ErrEmptyPipeline = errors.New("Empty pipeline, has no next element")

// Root is the root handler for pipelines.
type Root struct {
	next Pipeline
	opts *opts.Opts
	db   store.Store
}

// New creates a new Root instance
func New(opts *opts.Opts, db store.Store) *Root {
	return &Root{
		opts: opts,
		db:   db,
	}
}

// Publish ...
func (p *Root) Publish(m *model.Message) error {
	if p.next != nil {
		return p.next.Publish(m)
	}
	return nil
}

// AddNext ...
func (p *Root) AddNext(pe Pipeline) {
	p.next = pe
}

// Next ...
func (p *Root) Next() Pipeline {
	return p.next
}
