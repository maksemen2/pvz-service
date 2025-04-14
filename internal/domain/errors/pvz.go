package domainerrors

import "errors"

var (
	ErrUserNotModerator = errors.New("user is not moderator")       // Пользователь не является модератором
	ErrInvalidCity      = errors.New("invalid city provided")       // Недопустимый город
	ErrPVZAlreadyExists = errors.New("pvz already exists")          // Пункт выдачи уже существует
	ErrInvalidPage      = errors.New("invalid page provided")       // Недопустимая страница
	ErrInvalidLimit     = errors.New("invalid limit provided")      // Недопустимый лимит
	ErrInvalidDateRange = errors.New("invalid date range provided") // Недопустимый диапазон дат
	ErrInvalidStartDate = errors.New("invalid start date provided") // Недопустимая начальная дата
	ErrPVZNotFound      = errors.New("pvz not found")               // Пункт выдачи не найден
)
