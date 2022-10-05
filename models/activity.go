package models

import "time"

// Sample Activity used in this example to store Notificaiton data
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
