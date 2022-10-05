package server

import (
	"context"
)

// Custom error used specifically to indicate no account was found
type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "Resource not found"
}

const (
	DeactivateAccountSQL = `
	DELETE FROM accounts 
	WHERE resource_uuid=$1;
	`

	DeleteTokenSQL = `
	DELETE FROM tokens
	WHERE resource_uuid=$1
	`
)

// If given a deprovisioning request, delete the account's row and token entries
func (s *server) deprovisionRequest(ctx context.Context, uuid string) error {
	commandTag, err := s.db.Exec(ctx, DeactivateAccountSQL, uuid)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return &NotFoundError{}
	}

	_, err = s.db.Exec(ctx, DeleteTokenSQL, uuid)
	if err != nil {
		return err
	}

	return nil
}
