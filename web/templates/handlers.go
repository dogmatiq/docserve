package templates

import "github.com/dogmatiq/configkit"

type ListHandlersContext struct {
	Context

	Table []HandlerListRow
}

type HandlerListRow struct {
	HandlerKey      string
	HandlerName     string
	HandlerType     configkit.HandlerType
	HandlerTypeName string
	AppKey          string
	AppName         string
	MessageCount    string
}
