package models

import (
	"github.com/google/uuid"
	"time"
)

// Product - структура, представляющая товар в системе.
type Product struct {
	ID          uuid.UUID
	DateTime    time.Time
	Type        ProductType
	ReceptionID uuid.UUID
}

// AddProduct - структура, инкапсулирующая данные для добавления товара в приемку с указанием только айди пвз.
type AddProduct struct {
	ID       uuid.UUID
	DateTime time.Time
	Type     ProductType
	PVZID    uuid.UUID
}

type ProductType string

const (
	ProductTypeElectronics ProductType = "электроника"
	ProductTypeClothes     ProductType = "одежда"
	ProductTypeShoes       ProductType = "обувь"
)

func (p ProductType) Valid() bool {
	switch p {
	case ProductTypeElectronics, ProductTypeClothes, ProductTypeShoes:
		return true
	}

	return false
}

func (p ProductType) String() string {
	return string(p)
}
