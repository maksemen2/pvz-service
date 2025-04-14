package models

import (
	"github.com/google/uuid"
	"time"
)

type Reception struct {
	ID       uuid.UUID
	DateTime time.Time
	PVZID    uuid.UUID
	Status   ReceptionStatus // in_progress, close
}

type ReceptionStatus string

const (
	ReceptionStatusInProgress ReceptionStatus = "in_progress"
	ReceptionStatusClose      ReceptionStatus = "close"
)

func (r ReceptionStatus) Valid() bool {
	switch r {
	case ReceptionStatusInProgress, ReceptionStatusClose:
		return true
	}

	return false
}

func (r ReceptionStatus) String() string {
	return string(r)
}

type ReceptionWithProducts struct {
	Reception *Reception
	Products  []*Product
}
