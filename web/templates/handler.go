package templates

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
)

type HandlerListContext struct {
	Context

	TotalRepoCount int
	TotalAppCount  int
	Handlers       []HandlerRow
}

type HandlerRow struct {
	HandlerKey      string
	HandlerName     string
	HandlerType     configkit.HandlerType
	HandlerTypeName TypeName
	AppKey          string
	AppName         string
	MessageCount    int
}

type HandlerViewContext struct {
	Context

	HandlerKey      string
	HandlerName     string
	HandlerType     configkit.HandlerType
	HandlerTypeName TypeName
	AppKey          string
	AppName         string

	ConsumedMessages    []MessageRow
	ConsumedMessageRole message.Role
	ProducedMessages    []MessageRow
	ProducedMessageRole message.Role
	TimeoutMessages     []MessageRow
}
