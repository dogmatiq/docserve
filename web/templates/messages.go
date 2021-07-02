package templates

type ListMessagesContext struct {
	Context

	Table []MessageListRow
}

type MessageListRow struct {
	MessageTypeName string
	Role            string
	AppCount        int
	HandlerCount    int
}
