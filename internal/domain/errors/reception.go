package domainerrors

import "errors"

var (
	ErrNoOpenReceptions    = errors.New("no open receptions in this pvz")             // Нет открытых приемок в этом пункте выдачи
	ErrOpenReceptionExists = errors.New("open reception already exists for this pvz") // Уже существует открытая приемка для этого пункта выдачи
)
