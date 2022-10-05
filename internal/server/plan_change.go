package server

import (
	"context"
)

type PlanChangeRequest struct {
	PlanSlug string `json:"plan_slug"`
}

const (
	UpdatePlanSQL = `
	UPDATE accounts
	SET plan_slug=$2
	WHERE resource_uuid=$1;
	`
)

// If a user chooses to change their plan, DigitalOcean will send a Plan Change request
// with details of the new plan they are using
func (s *server) planChange(ctx context.Context, req *PlanChangeRequest, uuid string) error {
	commandTag, err := s.db.Exec(ctx, UpdatePlanSQL,
		uuid,
		req.PlanSlug,
	)

	if err != nil {
		return err
	} else if commandTag.RowsAffected() == 0 {
		return &NotFoundError{}
	}

	return nil
}
