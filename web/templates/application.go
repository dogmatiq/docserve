package templates

type ApplicationListContext struct {
	Context

	TotalRepoCount int
	Applications   []ApplicationRow
}

type ApplicationRow struct {
	AppKey       string
	AppName      string
	AppTypeName  TypeName
	HandlerCount int
	MessageCount int
}

type ApplicationViewContext struct {
	Context

	AppKey      string
	AppName     string
	AppTypeName TypeName

	Handlers []HandlerRow
	Messages []MessageRow
}
