// Package dto
package dto

import (
	"errors"
	"mime/multipart"
)

type CreateJobRequest struct {
	File   multipart.File
	Header *multipart.FileHeader
}

const MaxAudioSizeBytes = 100 * 1024 * 1024 // 100 MiB

func (r *CreateJobRequest) Validate() error {
	if r.File == nil || r.Header == nil {
		return errors.New("audio file is required")
	}

	if r.Header.Size == 0 {
		return errors.New("audio file is empty")
	}
	if r.Header.Size > MaxAudioSizeBytes {
		return errors.New("audio file is too large")
	}

	// TODO: Проверка формата
	return nil
}
