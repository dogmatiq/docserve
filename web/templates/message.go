package templates

type MessageListContext struct {
	Context

	TotalRepoCount    int
	TotalAppCount     int
	TotalHandlerCount int
	Messages          []MessageRow
}

type MessageRow struct {
	MessageTypeName TypeName
	MessageRole     string // could be multiple, comma separated
	AppCount        int
	HandlerCount    int
}

type MessageViewContext struct {
	Context

	MessageTypeName TypeName
	MessageRole     string // could be multiple, comma separated
}
