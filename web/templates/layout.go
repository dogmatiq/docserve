package templates

type Context struct {
	Title          string
	ActiveMenuItem MenuItem
}

type MenuItem string

const (
	ApplicationsMenuItem MenuItem = "applications"
	HandlersMenuItem              = "handlers"
	MessagesMenuItem              = "messages"
)
