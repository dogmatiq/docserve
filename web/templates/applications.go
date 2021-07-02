package templates

type ListApplicationsContext struct {
	Context

	Table []ApplicationListRow
}

type ApplicationListRow struct {
	AppKey       string
	AppName      string
	AppTypeName  string
	HandlerCount int
	MessageCount int
}
