package server

import (
	"context"
	"sample_app/models"
)

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "Resource not found"
}

const (
	DeactivateAccountSQL = `
	UPDATE accounts 
	SET status=$1
	WHERE resource_uuid=$2;
	`
)

// If given a deprovisioning request, update the status of the account to Deprovisioned
func (s *server) deprovisionRequest(ctx context.Context, uuid string) error {
	commandTag, err := s.db.Exec(ctx, DeactivateAccountSQL, models.Deprovisioned, uuid)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{}
	}
	return nil
}
