package models

import "time"

// Status of an account. Can be either active or suspsended.
type Status int64

const (
	Active Status = iota
	Suspended
)

// Sample Account structure used in this example
type Account struct {
	Id              int
	Name            string
	Email           string
	AppSlug         string
	PlanSlug        string
	ResourceUUID    string
	Language        string
	EmailPreference bool
	Source          string
	SourceId        string
	Status          Status
	LicenseKey      string
	CreatedAt       time.Time
	ModifiedAt      time.Time
}
