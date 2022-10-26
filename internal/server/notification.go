package server

import (
	"context"
	"encoding/json"
	"errors"
)

// These are some of the types of notifications DigitalOcean may send.
const (
	Suspended            = "resources.suspended"
	Reactivated          = "resources.reactivated"
	DeprovisioningFailed = "resources.deprovisioning.failed"
	Updated              = "resources.updated"

	InsertActivitySQL = `
	INSERT INTO activities (account_id, resource_uuid, type, title, body)
	VALUES ($1, $2, $3, $4, $5)
	`
)

type Notification struct {
	Type      string `json:"type" form:"type"`
	CreatedAt int    `json:"created_at" form:"created_at"`
	Payload   string `json:"payload" form:"payload"`
}

type SuspensionPayload struct {
	ResourceUUIDs []string `json:"resources_uuids"`
}

type ReactivatedPayload struct {
	ResourceUUIDs []string `json:"resources_uuids"`
}

type DeprovisioningFailedPayload struct {
	ResourceUUIDs []string `json:"resources_uuids"`
}

type UpdatedPayload struct {
	Resource ResourceState `json:"resource"`
	Plan     PlanState     `json:"plan"`
}

type ResourceState struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	State     string `json:"state"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

type PlanState struct {
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	CreatedAt   int    `json:"created_at"`
	UpdatedAt   int    `json:"updated_at"`
}

// Determine the type of a notification, and pass it to the relevant handler.
// Handlers would provide logic to respond to each type of notification. In our example,
// they simply record the notification as an Activity.
func (s *server) parseNotification(ctx context.Context, n *Notification) []error {
	var errs []error
	s.e.Logger.Info("Got notification of type " + n.Type)
	switch n.Type {
	case Suspended:
		errs = s.suspensionNotification(ctx, n)
	case Reactivated:
		errs = s.reactivationNotification(ctx, n)
	case DeprovisioningFailed:
		errs = s.deprovisionFailedNotification(ctx, n)
	case Updated:
		err := s.updateNotification(ctx, n)
		errs = append(errs, err)
	default:
		errs = append(errs, errors.New("unrecognized notification type"))
	}
	return errs
}

// Suspension notifications are used to communicate that a given account/user has
// been suspended for some reason (e.g. overdue billing)
func (s *server) suspensionNotification(ctx context.Context, n *Notification) []error {
	errs := []error{}
	payload := SuspensionPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, uuid := range payload.ResourceUUIDs {
		// Any logic needed to handle suspended users in your application would go here
		err = s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Reactivation notifications are used to communicate that a previously-suspended account/user
// is back in good standing.
func (s *server) reactivationNotification(ctx context.Context, n *Notification) []error {
	errs := []error{}
	payload := ReactivatedPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, uuid := range payload.ResourceUUIDs {
		// Any logic needed to handle reactivating users in your application would go here
		err = s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return nil
}

// Deprovisioning Failed notifications are used to inform you that a deprovisioning request
// for a given user failed.
func (s *server) deprovisionFailedNotification(ctx context.Context, n *Notification) []error {
	errs := []error{}
	payload := DeprovisioningFailedPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, uuid := range payload.ResourceUUIDs {
		// Logic to handle failed deprovisions would go here
		err = s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return nil
}

// Update notifications are sent when a user's information or plan changes.
func (s *server) updateNotification(ctx context.Context, n *Notification) error {
	payload := UpdatedPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		return err
	}

	// Logic to update a user's plan or state would go here
	err = s.writeNotification(ctx, n, payload.Resource.UUID)
	if err != nil {
		return err
	}

	return nil
}

// We write notifications to our Activities table for this example.
func (s *server) writeNotification(ctx context.Context, n *Notification, uuid string) error {
	var id int

	s.e.Logger.Info("Writing notification")
	err := s.db.QueryRow(ctx, GetAccountSQL, uuid).Scan(&id)
	if err != nil {
		s.e.Logger.Error("Error finding account id: " + err.Error())
		return err
	}

	_, err = s.db.Exec(ctx, InsertActivitySQL,
		id,
		uuid,
		"DigitalOcean",
		n.Type,
		n.Payload,
	)
	if err != nil {
		s.e.Logger.Error("Error writing notification: " + err.Error())
		return err
	}

	return nil
}
