package httpdto

import (
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"github.com/oapi-codegen/runtime/types"
)

func ToPVZResponse(pvz *models.PVZ) *PVZ {
	return &PVZ{
		City:             PVZCity(pvz.City),
		Id:               &pvz.ID,
		RegistrationDate: &pvz.RegistrationDate,
	}
}

func ModelToReceptionResponse(reception *models.Reception) *Reception {
	return &Reception{
		DateTime: reception.DateTime,
		Id:       &reception.ID,
		PvzId:    reception.PVZID,
		Status:   ReceptionStatus(reception.Status),
	}
}

func ModelToProductResponse(product *models.Product) *Product {
	return &Product{
		DateTime:    &product.DateTime,
		Id:          &product.ID,
		ReceptionId: product.ReceptionID,
		Type:        ProductType(product.Type),
	}
}

func ModelToReceptionWithProductsResponse(reception *models.ReceptionWithProducts) *ReceptionWithProductsResponse {
	products := make([]*Product, 0, len(reception.Products)) // лучше заранее выделить память

	for _, product := range reception.Products {
		products = append(products, ModelToProductResponse(product))
	}

	return &ReceptionWithProductsResponse{
		Reception: &Reception{
			DateTime: reception.Reception.DateTime,
			Id:       &reception.Reception.ID,
			PvzId:    reception.Reception.PVZID,
			Status:   ReceptionStatus(reception.Reception.Status),
		},
		Products: products,
	}
}

func ModelToPVZWithReceptionsResponse(pvz *models.PVZWithReceptions) *PVZWithReceptionsResponse {
	receptions := make([]*ReceptionWithProductsResponse, 0, len(pvz.Receptions))

	for _, reception := range pvz.Receptions {
		receptions = append(receptions, ModelToReceptionWithProductsResponse(reception))
	}

	return &PVZWithReceptionsResponse{
		PVZ: &PVZ{
			City:             PVZCity(pvz.PVZ.City),
			Id:               &pvz.PVZ.ID,
			RegistrationDate: &pvz.PVZ.RegistrationDate,
		},
		Receptions: receptions,
	}
}

func ModelToUserResponse(user *models.User) *User {
	return &User{
		Id:    &user.ID,
		Email: types.Email(user.Email),
		Role:  UserRole(user.Role),
	}
}
