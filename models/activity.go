package models

import "time"

type Activity struct {
	Id           int
	AccountId    int
	ResourceUUID string
	Type         int
	Title        string
	Body         string
	CreatedAt    time.Time
	ModifiedAt   time.Time
}
