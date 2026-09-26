package dto

import (
	"errors"
)

type GetJobRequest struct {
	ID int64
}

func (r *GetJobRequest) Validate() error {
	if r.ID <= 0 {
		return errors.New("id must be positive")
	}
	return nil
}
