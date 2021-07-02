package templates

import "github.com/dogmatiq/configkit/message"

type MessageListContext struct {
	Context

	Messages []MessageRow
}

type MessageRow struct {
	MessageTypeName TypeName
	Role            message.Role
	AppCount        int
	HandlerCount    int
}
