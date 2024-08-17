package correlator

import (
	"context"

	"github.com/dogmatiq/browser/model"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/minibus"
)

type Correlator struct {
	messages map[string]message.Roles
}

func (c *Correlator) Run(ctx context.Context) error {
	minibus.Subscribe[model.AppDiscovered](ctx)
	minibus.Ready(ctx)

	for m := range minibus.Inbox(ctx) {
		c.add(m.(model.AppDiscovered))
	}

	return nil
}

func (c *Correlator) add(m model.AppDiscovered) {
	if c.messages == nil {
		c.messages = map[string]message.Role{}
	}

	for _, m := range m.App.MessageNames().All() {

	}
}
