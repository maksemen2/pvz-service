package domainerrors

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")              // Пользователь уже существует
	ErrInvalidCredentials = errors.New("invalid credentials")              // Недопустимые учетные данные
	ErrUserNotFound       = errors.New("user not found")                   // Пользователь не найден
	ErrInvalidRole        = errors.New("invalid role")                     // Недопустимая роль
	ErrNotEnoughRights    = errors.New("not enough rights with role")      // Недостаточно прав с ролью (должна быть обёрнута)
	ErrPasswordTooLong    = errors.New("password too long, max length is") // Пароль слишком длинный (должна быть обёрнута)
)
