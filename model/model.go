package model

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
)

type Provider struct {
	ID uint `gorm:"primarykey"`

	Repositories []Repository
}

type Repository struct {
	ID uint `gorm:"primarykey"`

	ProviderID uint `gorm:"not null;uniqueIndex:provider_repository"`
	Provider   Provider

	ProviderRepositoryID string `gorm:"not null;uniqueIndex:provider_repository"`
	DisplayName          string `gorm:"not null"`

	Modules []Module
}

type Module struct {
	ID uint `gorm:"primarykey"`

	RepositoryID *uint
	Repository   *Repository

	Path string `gorm:"not null"`

	Applications []Application
	Messages     []Message
}

type Application struct {
	ID uint `gorm:"primarykey"`

	ModuleID uint `gorm:"not null"`
	Module   Module

	Key    string `gorm:"not null"`
	Name   string `gorm:"not null"`
	GoType string `gorm:"not null"`

	Handlers []Handler
}

type Handler struct {
	ID uint `gorm:"primarykey"`

	ApplicationID uint `gorm:"not null"`
	Application   Application

	Key         string                `gorm:"not null"`
	Name        string                `gorm:"not null"`
	GoType      string                `gorm:"not null"`
	HandlerType configkit.HandlerType `gorm:"not null"`
}

type HandlerMessage struct {
	HandlerID uint `gorm:"primarykey;autoIncrement:false"`
	Handler   Handler

	MessageID uint `gorm:"primarykey;autoIncrement:false"`
	Message   Message

	Role message.Role `gorm:"primarykey"`

	Produced bool
	Consumed bool
}

type Message struct {
	ID uint `gorm:"primarykey"`

	ModuleID *uint
	Module   *Module

	GoType string `gorm:"unique"`
}
