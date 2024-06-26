package server

import (
	"context"
	"errors"
	"fmt"
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

type Notification interface {
	GetType() string
	GetPayload() string
}

type SuspensionNotification struct {
	Type      string `json:"type"`
	CreatedAt int    `json:"created_at"`
	Payload   struct {
		ResourceUUIDs []string `json:"resources_uuids"`
	} `json:"payload"`
}

func (n *SuspensionNotification) GetType() string {
	return n.Type
}

func (n *SuspensionNotification) GetPayload() string {
	return fmt.Sprintf("%v", n.Payload)
}

type ReactivatedNotification struct {
	Type      string `json:"type"`
	CreatedAt int    `json:"created_at"`
	Payload   struct {
		ResourceUUIDs []string `json:"resources_uuids"`
	} `json:"payload"`
}

func (n *ReactivatedNotification) GetType() string {
	return n.Type
}

func (n *ReactivatedNotification) GetPayload() string {
	return fmt.Sprintf("%v", n.Payload)
}

type DeprovisioningFailedNotification struct {
	Type      string `json:"type"`
	CreatedAt int    `json:"created_at"`
	Payload   struct {
		ResourceUUIDs []string `json:"resources_uuids"`
	} `json:"payload"`
}

func (n *DeprovisioningFailedNotification) GetType() string {
	return n.Type
}

func (n *DeprovisioningFailedNotification) GetPayload() string {
	return fmt.Sprintf("%v", n.Payload)
}

type UpdatedNotification struct {
	Type      string `json:"type"`
	CreatedAt int    `json:"created_at"`
	Payload   struct {
		Resource ResourceState `json:"resource"`
		Plan     PlanState     `json:"plan"`
	} `json:"payload"`
}

func (n *UpdatedNotification) GetType() string {
	return n.Type
}

func (n *UpdatedNotification) GetPayload() string {
	return fmt.Sprintf("%v", n.Payload)
}

type ResourceState struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	State     string `json:"state"`
	CreatedAt struct {
		Seconds int `json:"seconds"`
	} `json:"created_at"`
	UpdatedAt struct {
		Seconds int `json:"seconds"`
	} `json:"updated_at"`
}

type PlanState struct {
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	CreatedAt   struct {
		Seconds int `json:"seconds"`
	} `json:"created_at"`
	UpdatedAt struct {
		Seconds int `json:"seconds"`
	} `json:"updated_at"`
}

// Determine the type of a notification, and pass it to the relevant handler.
// Handlers would provide logic to respond to each type of notification. In our example,
// they simply record the notification as an Activity.
func (s *server) parseNotification(ctx context.Context, n Notification) []error {
	var errs []error
	s.e.Logger.Info("Got notification of type " + n.GetType())
	switch n.GetType() {
	case Suspended:
		errs = s.suspensionNotification(ctx, n.(*SuspensionNotification))
	case Reactivated:
		errs = s.reactivationNotification(ctx, n.(*ReactivatedNotification))
	case DeprovisioningFailed:
		errs = s.deprovisionFailedNotification(ctx, n.(*DeprovisioningFailedNotification))
	case Updated:
		err := s.updateNotification(ctx, n.(*UpdatedNotification))
		if err != nil {
			errs = append(errs, err)
		}
	default:
		errs = append(errs, errors.New("unrecognized notification type"))
	}
	return errs
}

// Suspension notifications are used to communicate that a given account/user has
// been suspended for some reason (e.g. overdue billing)
func (s *server) suspensionNotification(ctx context.Context, n *SuspensionNotification) []error {
	errs := []error{}
	for _, uuid := range n.Payload.ResourceUUIDs {
		// Any logic needed to handle suspended users in your application would go here
		err := s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Reactivation notifications are used to communicate that a previously-suspended account/user
// is back in good standing.
func (s *server) reactivationNotification(ctx context.Context, n *ReactivatedNotification) []error {
	errs := []error{}
	for _, uuid := range n.Payload.ResourceUUIDs {
		// Any logic needed to handle reactivating users in your application would go here
		err := s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Deprovisioning Failed notifications are used to inform you that a deprovisioning request
// for a given user failed.
func (s *server) deprovisionFailedNotification(ctx context.Context, n *DeprovisioningFailedNotification) []error {
	errs := []error{}
	for _, uuid := range n.Payload.ResourceUUIDs {
		// Logic to handle failed deprovisions would go here
		err := s.writeNotification(ctx, n, uuid)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Update notifications are sent when a user's information or plan changes.
func (s *server) updateNotification(ctx context.Context, n *UpdatedNotification) error {
	// Logic to update a user's plan or state would go here
	err := s.writeNotification(ctx, n, n.Payload.Resource.UUID)
	if err != nil {
		return err
	}

	return nil
}

// We write notifications to our Activities table for this example.
func (s *server) writeNotification(ctx context.Context, n Notification, uuid string) error {
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
		n.GetType(),
		n.GetPayload(),
	)
	if err != nil {
		s.e.Logger.Error("Error writing notification: " + err.Error())
		return err
	}

	return nil
}
