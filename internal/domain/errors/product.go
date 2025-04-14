package domainerrors

import "errors"

var (
	ErrNoProductsInReception = errors.New("no products in this reception") // Нет продуктов в этой приемке
	ErrInvalidProductType    = errors.New("invalid product type")          // Недопустимый тип продукта
)
