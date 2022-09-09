package models

import "time"

type Status int64

const (
	Active Status = iota
	Suspended
	Deprovisioned
)

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
