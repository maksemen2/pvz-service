package models

import (
	"github.com/google/uuid"
	"time"
)

type PVZ struct {
	ID               uuid.UUID
	RegistrationDate time.Time
	City             CityType
}

type CityType string

const (
	CityTypeMoscow CityType = "Москва"
	CityTypeSPB    CityType = "Санкт-Петербург"
	CityTypeKazan  CityType = "Казань"
)

func (m CityType) Valid() bool {
	switch m {
	case CityTypeMoscow, CityTypeSPB, CityTypeKazan:
		return true
	}

	return false
}

func (m CityType) String() string {
	return string(m)
}

type PVZWithReceptions struct {
	PVZ        *PVZ
	Receptions []*ReceptionWithProducts
}
