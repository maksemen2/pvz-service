package domainerrors

import "errors"

var (
	ErrUnexpected = errors.New("unexpected error") // Непредвиденная ошибка сервиса
)
