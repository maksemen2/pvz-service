package databaseerrors

import "errors"

var (
	ErrUnexpected      = errors.New("unexpected database error") // Непредвиденная ошибка
	ErrUniqueViolation = errors.New("unique field violation")    // Нарушение уникальности
	ErrNoRows          = errors.New("no rows found")             // Строки не найдены
)
