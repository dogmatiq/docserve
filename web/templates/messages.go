package templates

type ListMessagesContext struct {
	Context

	Table []MessageListRow
}

type MessageListRow struct {
	MessageTypeName TypeName
	Role            string
	AppCount        int
	HandlerCount    int
}
