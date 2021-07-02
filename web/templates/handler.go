package templates

import "github.com/dogmatiq/configkit"

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
