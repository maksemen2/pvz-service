// Пакет commonerrors содержит функции для генерации стандартных ошибок, возвращаемых в HTTP хендлерах.
// Функции возвращают сгенерированную DTO модель httpdto.Error
package commonerrors

import "github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"

func Unauthorized() httpdto.Error {
	return httpdto.Error{Message: "unauthorized"}
}

func InvalidCredentials() httpdto.Error {
	return httpdto.Error{Message: "invalid email or password"}
}

func BadRequest(message string) httpdto.Error {
	return httpdto.Error{Message: message}
}

func Internal() httpdto.Error {
	return httpdto.Error{Message: "internal server error"}
}

func Forbidden() httpdto.Error {
	return httpdto.Error{Message: "forbidden"}
}
