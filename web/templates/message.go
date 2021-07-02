package templates

type MessageListContext struct {
	Context

	Messages []MessageRow
}

type MessageRow struct {
	MessageTypeName TypeName
	Role            string
	AppCount        int
	HandlerCount    int
}
