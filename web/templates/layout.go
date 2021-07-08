package templates

import "github.com/google/go-github/v35/github"

type Context struct {
	Title          string
	User           *github.User
	ActiveMenuItem MenuItem
}

type MenuItem string

const (
	ApplicationsMenuItem MenuItem = "applications"
	HandlersMenuItem     MenuItem = "handlers"
	MessagesMenuItem     MenuItem = "messages"
)
