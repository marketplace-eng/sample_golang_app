package server

import (
	"context"
	"encoding/json"
	"errors"
)

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
	Type      string `json:"type"`
	CreatedAt int    `json:"created_at"`
	Payload   string `json:"payload"`
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

func (s *server) parseNotification(ctx context.Context, n *Notification) []error {
	var errs []error
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

func (s *server) suspensionNotification(ctx context.Context, n *Notification) []error {
	errs := []error{}
	payload := SuspensionPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, uuid := range payload.ResourceUUIDs {
		// Logic to suspend a user would go here
		err = s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *server) reactivationNotification(ctx context.Context, n *Notification) []error {
	errs := []error{}
	payload := ReactivatedPayload{}
	err := json.Unmarshal([]byte(n.Payload), &payload)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, uuid := range payload.ResourceUUIDs {
		// Logic to reactivate a user would go here
		err = s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return nil
}

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

func (s *server) writeNotification(ctx context.Context, n *Notification, uuid string) error {
	var id int

	err := s.db.QueryRow(ctx, GetAccountSQL, uuid).Scan(&id)
	if err != nil {
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
		return err
	}

	return nil
}
